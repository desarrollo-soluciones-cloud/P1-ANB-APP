import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { VideoService } from '../../../services/video.service';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { environment } from '../../../../environments/environment';

@Component({
  selector: 'app-video-detail',
  standalone: true,
  imports: [CommonModule, RouterModule, MatCardModule, MatButtonModule, MatIconModule],
  templateUrl: './detail.component.html',
  styleUrls: ['./detail.component.scss']
})
export class VideoDetailComponent implements OnInit {
  video: any = null;

  constructor(private route: ActivatedRoute, private videoService: VideoService) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.videoService.getVideoById(id).subscribe({
        next: (res) => {
          this.video = res;
        },
        error: (err) => {
          console.error('Error fetching video detail', err);
        }
      });
    }
  }

  normalizeUrl(url: any): string | null {
    if (!url) return null;
    const s = String(url);
    if (s.startsWith('http://') || s.startsWith('https://')) return s;
    const base = environment.apiUrl?.replace(/\/$/, '');
    return base ? `${base}${s.startsWith('/') ? '' : '/'}${s}` : s;
  }

  getVideoUrl(video: any, type: 'processed' | 'original'): string | null {
    const url = type === 'processed' ? video.processed_url : video.original_url;
    if (!url) return null;
    
    // URLs de S3 o completas se devuelven directamente
    if (url.startsWith('http://') || url.startsWith('https://')) {
      return url;
    }
    
    // Backward compatibility: Para rutas locales /uploads
    const apiBase = environment.apiUrl || 'http://localhost:9090';
    if (url.startsWith('/uploads') || url.startsWith('uploads')) {
      const cleanPath = url.startsWith('/') ? url : '/' + url;
      return `${apiBase}${cleanPath}`;
    }
    
    return `${apiBase}/${url}`;
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
