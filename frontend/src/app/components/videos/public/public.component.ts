import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { VideoService } from '../../../services/video.service';
import { environment } from '../../../../environments/environment';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { RouterModule } from '@angular/router';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';

@Component({
  selector: 'app-public-videos',
  standalone: true,
  imports: [CommonModule, MatCardModule, MatButtonModule, MatIconModule, MatListModule, MatProgressSpinnerModule, RouterModule],
  templateUrl: './public.component.html',
  styleUrls: ['./public.component.scss']
})
export class PublicVideosComponent implements OnInit {
  videos: any[] = [];
  loading = false;
  error?: string;

  constructor(private videoService: VideoService) {}

  ngOnInit(): void {
    this.loadPublic();
  }

  loadPublic() {
    this.loading = true;
    this.videoService.listPublicVideos().subscribe({
      next: (res) => {
        this.loading = false;
        if (Array.isArray(res)) this.videos = res;
        else if (res?.data && Array.isArray(res.data)) this.videos = res.data;
        else this.videos = res?.videos || [];
        
        // Debug: Log the data structure to see what info we have
        console.log('Videos públicos datos completos:', this.videos);
        if (this.videos.length > 0) {
          console.log('Primer video estructura:', this.videos[0]);
        }

        // Load user information for each video
        this.loadUsersInfo();
      },
      error: (err) => {
        this.loading = false;
        this.error = err?.error?.message || err?.message || 'Error cargando videos públicos.';
      }
    });
  }

  loadUsersInfo() {
    if (this.videos.length === 0) return;

    // Get unique user IDs
    const userIds = [...new Set(this.videos.map(video => video.user_id).filter(id => id))];
    console.log('User IDs a buscar:', userIds);

    if (userIds.length === 0) return;

    // Try multiple endpoint patterns to get user info
    const userRequests = userIds.map(userId => {
      return this.videoService.getUserById(userId).pipe(
        catchError(() => {
          // If first endpoint fails, try alternative
          return this.videoService.getUserProfile(userId).pipe(
            catchError(() => {
              console.log(`No se pudo obtener info para user_id: ${userId}`);
              return of(null);
            })
          );
        })
      );
    });

    forkJoin(userRequests).subscribe({
      next: (users) => {
        console.log('Información de usuarios obtenida:', users);
        // Map user info to videos
        users.forEach((user, index) => {
          if (user) {
            const userId = userIds[index];
            this.videos.forEach(video => {
              if (video.user_id === userId) {
                video.user_info = user;
              }
            });
          }
        });
        console.log('Videos con información de usuario:', this.videos);
      },
      error: (err) => {
        console.log('Error obteniendo información de usuarios:', err);
      }
    });
  }

  normalizeUrl(url: string) {
    if (!url) return null;
  if (url.startsWith('http://') || url.startsWith('https://')) return url;
  const base = environment.apiUrl?.replace(/\/$/, '');
  return base ? `${base}${url.startsWith('/') ? '' : '/'}${url}` : url;
  }

  // Vote action triggered by the UI
  vote(video: any) {
    if (!this.isAuthenticated()) {
      this.error = 'Debes iniciar sesión para votar.';
      return;
    }
    const id = this.getVideoId(video);
    this.videoService.voteVideo(id).subscribe({
      next: (res) => {
        // increment local counter and optionally mark as voted
        video.votes = this.getVoteCount(video) + 1;
        video._voted = true;
        this.error = undefined;
      },
      error: (err) => {
        if (err?.status === 400) {
          this.error = 'Ya has votado por este video.';
        } else if (err?.status === 401) {
          this.error = 'No autenticado.';
        } else if (err?.status === 404) {
          this.error = 'Video no encontrado.';
        } else {
          this.error = err?.error?.error || err?.error?.message || err?.message || 'Error registrando el voto.';
        }
      }
    });
  }

  unvote(video: any) {
    if (!this.isAuthenticated()) {
      this.error = 'Debes iniciar sesión para retirar el voto.';
      return;
    }
    const id = this.getVideoId(video);
    this.videoService.unvoteVideo(id).subscribe({
      next: (res) => {
        video.votes = Math.max(0, this.getVoteCount(video) - 1);
        video._voted = false;
        this.error = undefined;
      },
      error: (err) => {
        if (err?.status === 404) {
          this.error = 'No existe tu voto para este video.';
        } else if (err?.status === 401) {
          this.error = 'No autenticado.';
        } else {
          this.error = err?.error?.error || err?.error?.message || err?.message || 'Error retirando el voto.';
        }
      }
    });
  }

  isAuthenticated(): boolean {
    // simple check, AuthService would be better but keep local to avoid extra import
    return !!localStorage.getItem('token');
  }

  getErrorTitle(): string {
    if (this.error?.includes('autenticado') || this.error?.includes('login')) {
      return 'Necesitas iniciar sesión';
    }
    return 'Error al cargar videos';
  }

  getAuthorName(video: any): string {
    // First try user_info from our API call
    if (video.user_info) {
      return video.user_info.first_name && video.user_info.last_name 
        ? `${video.user_info.first_name} ${video.user_info.last_name}`
        : video.user_info.name || 
          video.user_info.username || 
          video.user_info.email ||
          'Autor Desconocido';
    }
    
    // Fallback to direct fields
    return video.user_name || 
           video.author_name || 
           video.author?.name || 
           video.author?.username || 
           video.user?.name || 
           video.user?.username || 
           video.username || 
           `Usuario ${video.user_id || 'Anónimo'}`;
  }

  getAuthorCity(video: any): string | null {
    // First try user_info from our API call
    if (video.user_info?.city) {
      return video.user_info.city;
    }
    
    // Fallback to direct fields
    return video.user_city || 
           video.author_city || 
           video.author?.city || 
           video.user?.city || 
           video.city || 
           null;
  }

  getAuthorCountry(video: any): string | null {
    // Check if user_info has country information
    if (video.user_info?.country) {
      return video.user_info.country;
    }
    
    return video.user_country || 
           video.author_country || 
           video.author?.country || 
           video.user?.country || 
           video.country || 
           null;
  }

  getVideoId(video: any): string {
    return video.video_id || video.id || '';
  }

  getVoteCount(video: any): number {
    return video.votes || video.vote_count || video.voteCount || 0;
  }
}
