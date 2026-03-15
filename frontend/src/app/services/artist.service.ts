import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Artist, FilterRequest } from '../models/artist.model';

@Injectable({ providedIn: 'root' })
export class ArtistService {
  private base = '/api/artists';

  constructor(private http: HttpClient) {}

  getAll(): Observable<Artist[]> {
    return this.http.get<Artist[]>(this.base);
  }

  // Query backend for initial load readiness
  getLoaded(): Observable<{ loaded: boolean }> {
    return this.http.get<{ loaded: boolean }>('/api/loaded');
  }

  getById(id: number): Observable<Artist> {
    return this.http.get<Artist>(`${this.base}/${id}`);
  }

  create(artist: Partial<Artist>) {
    return this.http.post<Artist>(this.base, artist);
  }

  update(id: number, artist: Partial<Artist>) {
    return this.http.put<Artist>(`${this.base}/${id}`, artist);
  }

  delete(id: number) {
    return this.http.delete(`${this.base}/${id}`);
  }

  search(query: string) {
    return this.http.post<Artist[]>('/api/search', { query });
  }

  filter(req: FilterRequest): Observable<Artist[]> {
    return this.http.post<Artist[]>(`${this.base}/filter`, req);
  }
}
