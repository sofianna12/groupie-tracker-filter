import { Component, OnInit, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ArtistService } from '../../services/artist.service';
import { NgForOf } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Artist, FilterRequest } from '../../models/artist.model';

@Component({
  selector: 'app-artists',
  standalone: true,
  imports: [CommonModule, RouterModule, NgForOf, FormsModule],
  templateUrl: './artists.component.html',
  styleUrls: ['./artists.component.scss']
})
export class ArtistsComponent implements OnInit {
  // full data set (source of truth)
  allArtists: Artist[] = [];
  // filtered + sorted list used for display & pagination
  filteredArtists: Artist[] = [];

  // UI state
  loading = false;
  error: string | null = null;
  serverSearching = false;
  showAddForm = false;
  editingArtist: Artist | null = null;

  // Form data
  newArtist = {
    name: '',
    image: '',
    members: '',
    firstAlbum: '',
    locations: ''
  };

  // search & sort
  searchQuery = '';
  sortField: 'name' | 'creationDate' = 'name';
  sortDir: 'asc' | 'desc' = 'asc';

  // filter panel
  showFilters = false;
  filterActive = false;
  filters: FilterRequest = {
    creationDateFrom: 0,
    creationDateTo: 0,
    firstAlbumFrom: 0,
    firstAlbumTo: 0,
    membersCount: [],
    locations: []
  };

  // dynamic checkbox options derived from loaded data
  availableMemberCounts: number[] = [];
  availableLocations: string[] = [];
  selectedMemberCounts: Set<number> = new Set();
  selectedLocations: Set<string> = new Set();

  // slider bounds for creation date (derived from loaded data)
  creationDateMin = 1950;
  creationDateMax = new Date().getFullYear();

  // Pagination state
  currentPage = 1;
  pageSize = 10; // default
  pageSizeOptions = [10, 20, 50];

  constructor(private artistService: ArtistService, private cdr: ChangeDetectorRef) {}

  // Small inline SVG placeholder (data URI) used when remote images fail or URL missing
  readonly placeholderDataUri = `data:image/svg+xml;utf8,` + encodeURIComponent(
    `<svg xmlns='http://www.w3.org/2000/svg' width='160' height='160' viewBox='0 0 160 160'>` +
      `<rect width='100%' height='100%' fill='%23f0f2f5'/>` +
      `<text x='50%' y='50%' dominant-baseline='middle' text-anchor='middle' fill='%23999' font-family='Arial,Helvetica,sans-serif' font-size='14'>No image</text>` +
    `</svg>`
  );

  ngOnInit(): void {
    this.fetchAll();
  }

  fetchAll() {
    this.loading = true;
    this.error = null;
    this.artistService.getAll().subscribe({
      next: (data) => {
        console.log('Received artists:', data?.length || 0);
        this.allArtists = data || [];
        this.loading = false;
        this.buildFilterOptions();
        this.applyFilterSort();
        this.cdr.detectChanges();
        // If initial result is empty, start polling backend readiness
        if (!this.allArtists.length) {
          this.startPollingForLoad();
        }
      },
      error: (err) => {
        console.error('Error fetching artists:', err);
        this.error = err?.message || String(err);
        this.loading = false;
      }
    });
  }

  private serverPollTimer: any = null;
  private pollAttempts = 0;
  pollingForLoad = false;

  startPollingForLoad() {
    if (this.serverPollTimer) return; // already polling
    this.pollAttempts = 0;
    this.pollingForLoad = true;
    this.serverPollTimer = setInterval(() => {
      this.pollAttempts++;
      this.artistService.getLoaded().subscribe({
        next: (res) => {
          if (res?.loaded) {
            // stop polling and refetch data
            clearInterval(this.serverPollTimer);
            this.serverPollTimer = null;
            this.pollingForLoad = false;
            this.fetchAll();
          } else if (this.pollAttempts > 30) {
            // give up after ~1 minute
            clearInterval(this.serverPollTimer);
            this.serverPollTimer = null;
            this.pollingForLoad = false;
          }
        },
        error: () => {
          // ignore and keep polling
          if (this.pollAttempts > 30) {
            clearInterval(this.serverPollTimer);
            this.serverPollTimer = null;
              this.pollingForLoad = false;
          }
        }
      });
    }, 2000);
  }

  onSearch() {
    const q = (this.searchQuery || '').trim();
    if (!q) {
      // empty query — fetch all (or use local copy)
      this.fetchAll();
      return;
    }

    // Immediate client-side filter for snappy UX
    this.applyFilterSort();

    // Then request authoritative server-side search in background and update when ready
    this.serverSearching = true;
    this.artistService.search(q).subscribe({
      next: (data) => {
        // replace base data with server results and re-apply filter/sort
        this.allArtists = data || [];
        this.applyFilterSort();
        this.serverSearching = false;
      },
      error: (err) => {
        this.error = err?.message || String(err);
        this.serverSearching = false;
      }
    });
  }

  // Debounced client-side search handler (fast UI). Server search is only triggered
  // when user explicitly clicks Search or presses Enter (onSearch()).
  private searchDebounceTimer: any = null;

  onSearchInput(value: string) {
    this.searchQuery = value;
    if (this.searchDebounceTimer) {
      clearTimeout(this.searchDebounceTimer);
    }
    // debounce 250ms
    this.searchDebounceTimer = setTimeout(() => {
      this.applyFilterSort();
      this.searchDebounceTimer = null;
    }, 250);
  }

  applyFilterSort() {
    const q = (this.searchQuery || '').trim().toLowerCase();

    // client-side filter (works even if server search used)
    this.filteredArtists = this.allArtists.filter(a => a.name.toLowerCase().includes(q));

    // sort
    const dir = this.sortDir === 'asc' ? 1 : -1;
    this.filteredArtists.sort((a, b) => {
      if (this.sortField === 'name') {
        return dir * a.name.localeCompare(b.name);
      }
      // sort by parsed date from firstAlbum (best source of full date), fallback to creationDate
      const da = this.parseAlbumDate(a) || Number(a.creationDate || 0);
      const db = this.parseAlbumDate(b) || Number(b.creationDate || 0);
      return dir * (da - db);
    });

    // reset to first page after filter/sort
    this.currentPage = 1;
  }

  // Try to extract a timestamp (ms) from the artist.firstAlbum field.
  // Prefer ISO date parse, then YYYY-MM-DD, then 4-digit year fallback.
  private parseAlbumDate(a: Artist): number {
    if (!a || !a.firstAlbum) return 0;
    const s = a.firstAlbum.trim();
    // Try Date.parse first (handles ISO strings)
    const p = Date.parse(s);
    if (!isNaN(p)) return p;

    // Look for YYYY-MM-DD
    const isoMatch = s.match(/(\d{4}-\d{2}-\d{2})/);
    if (isoMatch) {
      const t = Date.parse(isoMatch[1]);
      if (!isNaN(t)) return t;
    }

    // Look for 4-digit year
    const yearMatch = s.match(/(\d{4})/);
    if (yearMatch) {
      const y = Number(yearMatch[1]);
      if (y > 0) return Date.UTC(y, 0, 1);
    }

    return 0;
  }

  totalPages(totalItems: number): number {
    return Math.max(1, Math.ceil(totalItems / this.pageSize));
  }

  changePage(page: number, totalItems?: number) {
    if (page < 1) page = 1;
    if (totalItems !== undefined) {
      const tp = this.totalPages(totalItems);
      if (page > tp) page = tp;
    }
    this.currentPage = page;
  }

  changePageSize(value: string | number, totalItems?: number) {
    const newSize = Number(value);
    if (!isFinite(newSize) || newSize <= 0) return;
    this.pageSize = newSize;
    // clamp current page
    if (totalItems !== undefined) {
      const tp = this.totalPages(totalItems);
      if (this.currentPage > tp) this.currentPage = tp;
    } else {
      this.currentPage = 1;
    }
  }

  pages(totalItems: number): number[] {
    const tp = this.totalPages(totalItems);
    return Array.from({ length: tp }, (_, i) => i + 1);
  }

  // toggle sort field/dir
  setSort(field: 'name' | 'creationDate') {
    if (this.sortField === field) {
      this.sortDir = this.sortDir === 'asc' ? 'desc' : 'asc';
    } else {
      this.sortField = field;
      this.sortDir = 'asc';
    }
    this.applyFilterSort();
  }

  clearSearch() {
    this.searchQuery = '';
    this.applyFilterSort();
  }

  // Rebuild checkbox options and slider bounds whenever artists are loaded
  private buildFilterOptions() {
    const counts = new Set<number>();
    const locs = new Set<string>();
    let minYear = 9999;
    let maxYear = 0;

    for (const a of this.allArtists) {
      counts.add(a.members.length);
      for (const l of (a.locations || [])) {
        locs.add(l);
      }
      if (a.creationDate > 0) {
        if (a.creationDate < minYear) minYear = a.creationDate;
        if (a.creationDate > maxYear) maxYear = a.creationDate;
      }
    }

    this.availableMemberCounts = Array.from(counts).sort((a, b) => a - b);
    this.availableLocations = Array.from(locs).sort();

    if (maxYear > 0) {
      this.creationDateMin = minYear;
      this.creationDateMax = maxYear;
    }

    // initialise sliders to full range on first load (only if not already set)
    if (!this.filters.creationDateFrom) this.filters.creationDateFrom = this.creationDateMin;
    if (!this.filters.creationDateTo)   this.filters.creationDateTo   = this.creationDateMax;
  }

  // Clamp sliders so From never exceeds To
  onCreationFromChange() {
    if (this.filters.creationDateFrom > this.filters.creationDateTo) {
      this.filters.creationDateTo = this.filters.creationDateFrom;
    }
  }

  onCreationToChange() {
    if (this.filters.creationDateTo < this.filters.creationDateFrom) {
      this.filters.creationDateFrom = this.filters.creationDateTo;
    }
  }

  toggleMemberCount(n: number) {
    if (this.selectedMemberCounts.has(n)) {
      this.selectedMemberCounts.delete(n);
    } else {
      this.selectedMemberCounts.add(n);
    }
  }

  toggleLocation(loc: string) {
    if (this.selectedLocations.has(loc)) {
      this.selectedLocations.delete(loc);
    } else {
      this.selectedLocations.add(loc);
    }
  }

  applyFilters() {
    // Sliders are always set; only treat as active if they don't span the full range
    const creationActive = this.filters.creationDateFrom > this.creationDateMin ||
                           this.filters.creationDateTo   < this.creationDateMax;

    const req: FilterRequest = {
      creationDateFrom: creationActive ? this.filters.creationDateFrom : 0,
      creationDateTo:   creationActive ? this.filters.creationDateTo   : 0,
      firstAlbumFrom: this.filters.firstAlbumFrom || 0,
      firstAlbumTo:   this.filters.firstAlbumTo   || 0,
      membersCount: Array.from(this.selectedMemberCounts),
      locations:    Array.from(this.selectedLocations)
    };

    const isEmpty = !req.creationDateFrom && !req.creationDateTo &&
      !req.firstAlbumFrom && !req.firstAlbumTo &&
      req.membersCount.length === 0 && req.locations.length === 0;

    if (isEmpty) {
      this.filterActive = false;
      this.fetchAll();
      return;
    }

    this.filterActive = true;
    this.loading = true;
    this.artistService.filter(req).subscribe({
      next: (data: Artist[]) => {
        this.allArtists = data || [];
        this.loading = false;
        this.applyFilterSort();
        this.cdr.detectChanges();
      },
      error: (err: { message?: string }) => {
        this.error = err?.message || String(err);
        this.loading = false;
      }
    });
  }

  clearFilters() {
    this.filters = {
      creationDateFrom: 0,
      creationDateTo: 0,
      firstAlbumFrom: 0,
      firstAlbumTo: 0,
      membersCount: [],
      locations: []
    };
    this.selectedMemberCounts = new Set();
    this.selectedLocations = new Set();
    this.filterActive = false;
    this.fetchAll();
  }

  // Add new artist
  toggleAddForm() {
    this.showAddForm = !this.showAddForm;
    if (!this.showAddForm) {
      this.resetForm();
    }
  }

  resetForm() {
    this.newArtist = {
      name: '',
      image: '',
      members: '',
      firstAlbum: '',
      locations: ''
    };
    this.editingArtist = null;
  }

  submitArtist() {
    const members = this.newArtist.members.split(',').map(m => m.trim()).filter(m => m);
    const locations = this.newArtist.locations.split(',').map(l => l.trim()).filter(l => l);
    const artist = {
      name: this.newArtist.name,
      image: this.newArtist.image,
      members,
      creationDate: new Date().getFullYear(),
      firstAlbum: this.newArtist.firstAlbum,
      locations,
      dates: []
    };

    if (this.editingArtist) {
      // Update existing
      this.artistService.update(this.editingArtist.id, artist).subscribe({
        next: () => {
          this.fetchAll();
          this.showAddForm = false;
          this.resetForm();
        },
        error: (err) => {
          this.error = err?.message || String(err);
        }
      });
    } else {
      // Create new
      this.artistService.create(artist).subscribe({
        next: () => {
          this.fetchAll();
          this.showAddForm = false;
          this.resetForm();
        },
        error: (err) => {
          this.error = err?.message || String(err);
        }
      });
    }
  }

  editArtist(artist: Artist) {
    this.editingArtist = artist;
    this.newArtist = {
      name: artist.name,
      image: artist.image,
      members: artist.members.join(', '),
      firstAlbum: artist.firstAlbum,
      locations: (artist.locations || []).join(', ')
    };
    this.showAddForm = true;
  }

  deleteArtist(id: number, name: string) {
    if (!confirm(`Delete artist "${name}"?`)) return;
    
    this.artistService.delete(id).subscribe({
      next: () => {
        this.fetchAll();
      },
      error: (err) => {
        this.error = err?.message || String(err);
      }
    });
  }

  // Ensure the URL is usable in an <img src>. If missing a scheme, assume https.
  normalizeUrl(url?: string): string {
    if (!url) return this.placeholderDataUri;
    const s = url.trim();
    // If it already looks like an http(s) URL, return as-is
    if (/^https?:\/\//i.test(s)) return s;
    // If it starts with '//' (protocol-relative), prefix https:
    if (/^\/\//.test(s)) return 'https:' + s;
    // Otherwise, try to prepend https:// (best-effort)
    return 'https://' + s;
  }

  onImgError(ev: Event) {
    const img = ev.target as HTMLImageElement;
    if (!img) return;
    // avoid infinite loop if the placeholder also fails
    if (img.dataset['fallback'] === '1') return;
    img.dataset['fallback'] = '1';
    img.src = this.placeholderDataUri;
  }
}
