import { Routes } from '@angular/router';
import { ArtistsComponent } from './components/artists/artists.component';
import { ArtistDetailComponent } from './components/artist-detail/artist-detail.component';

export const routes: Routes = [
	{ path: '', redirectTo: '/artists', pathMatch: 'full' },
	{ path: 'artists', component: ArtistsComponent },
	{ path: 'artists/:id', component: ArtistDetailComponent }
];
