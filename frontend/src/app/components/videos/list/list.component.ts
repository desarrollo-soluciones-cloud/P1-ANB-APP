import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatListModule } from '@angular/material/list';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTooltipModule } from '@angular/material/tooltip';
import { VideoService } from '../../../services/video.service';
import { RouterModule } from '@angular/router';
import { environment } from '../../../../environments/environment';

@Component({
  selector: 'app-list-videos',
  standalone: true,
  imports: [CommonModule, MatListModule, MatCardModule, MatButtonModule, MatIconModule, MatProgressSpinnerModule, MatTooltipModule, RouterModule],
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.scss']
})
export class ListVideosComponent implements OnInit {
  videos: any[] = [];
  loading = false;
  error?: string;

  constructor(private videoService: VideoService) { }

  ngOnInit(): void {
    this.loadVideos();
  }

  loadVideos() {
    this.loading = true;
    this.videoService.listMyVideos().subscribe({
      next: (res) => {
        this.loading = false;
        if (Array.isArray(res)) this.videos = res;
        else if (res?.data && Array.isArray(res.data)) this.videos = res.data;
        else this.videos = res?.videos || [];
      },
      error: (err) => {
        this.loading = false;
        this.error = err?.error?.message || err?.message || 'Error cargando videos.';
      }
    });
  }

  processedUrl(video: any) {
    return this.normalizeUrl(video.processed_url || video.processedUrl);
  }

  originalUrl(video: any) {
    return this.normalizeUrl(video.original_url || video.originalUrl);
  }

  videoId(video: any) {
    return video.video_id || video.id;
  }

  normalizeUrl(url: string | undefined | null) {
    if (!url) return null;
    const s = String(url);
    if (s.startsWith('http://') || s.startsWith('https://')) return s;
    // Para uploads, usar ruta relativa (manejada por nginx proxy)
    if (s.startsWith('uploads/')) {
      return `/${s}`;
    }
    // Para otras rutas, usar apiUrl
    const base = environment.apiUrl?.replace(/\/$/, '');
    return base ? `${base}${s.startsWith('/') ? '' : '/'}${s}` : s;
  }

  getVideoUrl(video: any, type: 'processed' | 'original'): string | null {
    const url = type === 'processed' ? this.processedUrl(video) : this.originalUrl(video);
    if (!url) return null;
    
    // Usar ruta relativa para evitar CORS (manejada por nginx proxy)
    if (url.startsWith('/uploads') || url.startsWith('uploads')) {
      const cleanPath = url.startsWith('/') ? url : '/' + url;
      return cleanPath; // Ruta relativa que nginx proxy manejará
    }
    
    return url;
  }

  onDelete(video: any) {
    const vid = this.videoId(video);
    const confirmed = confirm(`¿Eliminar video "${video.title}"? Esta acción no se puede deshacer.`);
    if (!confirmed) return;

    this.videoService.deleteVideo(vid).subscribe({
      next: (res) => {
        // refresh list
        this.loadVideos();
      },
      error: (err) => {
        // Show a clearer message for 400 Bad Request (backend rule prevents deletion)
        if (err?.status === 400) {
          this.error = err?.error?.error || 'El video no puede eliminarse por su estado (procesado/publicado).';
        } else if (err?.status === 403) {
          this.error = 'No tienes permisos para eliminar este video.';
        } else if (err?.status === 404) {
          this.error = 'Video no encontrado.';
        } else {
          this.error = err?.error?.message || err?.message || 'Error eliminando el video.';
        }
      }
    });
  }

  isDeletable(video: any) {
    const s = (video?.status || '').toString().trim().toLowerCase();
    return s === 'uploaded';
  }

  getStatusClass(status: string): string {
    const s = (status || '').toLowerCase();
    if (s === 'processed') return 'status-processed';
    if (s === 'processing') return 'status-processing';
    return 'status-uploaded';
  }

  getStatusIcon(status: string): string {
    const s = (status || '').toLowerCase();
    if (s === 'processed') return 'check_circle';
    if (s === 'processing') return 'autorenew';
    return 'upload';
  }

  getStatusText(status: string): string {
    const s = (status || '').toLowerCase();
    if (s === 'processed') return 'Procesado';
    if (s === 'processing') return 'Procesando';
    return 'Subido';
  }
}
