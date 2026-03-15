export interface Artist {
  id: number;
  name: string;
  image: string;
  firstAlbum: string;
  creationDate: number;
  members: string[];
  locations: string[];
  dates: string[];
  datesLocations: { [location: string]: string[] };
}

export interface FilterRequest {
  creationDateFrom: number;
  creationDateTo: number;
  firstAlbumFrom: number;
  firstAlbumTo: number;
  membersCount: number[];
  locations: string[];
}
