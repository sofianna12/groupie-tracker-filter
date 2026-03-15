import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { ArtistService } from '../../services/artist.service';
import { switchMap } from 'rxjs/operators';
import { Observable } from 'rxjs';
import { Artist } from '../../models/artist.model';

@Component({
  selector: 'app-artist-detail',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './artist-detail.component.html',
  styleUrls: ['./artist-detail.component.scss']
})
export class ArtistDetailComponent {
  artist$!: Observable<Artist>;

  constructor(private route: ActivatedRoute, private artistService: ArtistService) {
    this.artist$ = this.route.paramMap.pipe(
      switchMap(params => {
        const id = Number(params.get('id'));
        return this.artistService.getById(id);
      })
    );
  }
}
