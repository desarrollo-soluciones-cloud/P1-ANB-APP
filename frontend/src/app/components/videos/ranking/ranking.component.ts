import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatCardModule } from '@angular/material/card';
import { RouterModule } from '@angular/router';
import { VideoService } from '../../../services/video.service';

@Component({
  selector: 'app-ranking',
  standalone: true,
  imports: [CommonModule, MatTableModule, MatButtonModule, MatIconModule, MatProgressSpinnerModule, MatCardModule, RouterModule],
  templateUrl: './ranking.component.html',
  styleUrls: ['./ranking.component.scss']
})
export class RankingComponent implements OnInit {
  rankings: any[] = [];
  loading = false;
  error?: string;
  columns = ['position', 'title', 'author_name', 'votes'];

  constructor(private videoService: VideoService) {}

  ngOnInit(): void {
    this.loadRankings();
  }

  loadRankings() {
    this.loading = true;
    this.videoService.listRankings().subscribe({
      next: (res) => {
        this.loading = false;
        const arr = Array.isArray(res) ? res : res?.data || [];
        // Normalize different possible response shapes to a consistent object
        this.rankings = arr.map((r: any) => ({
          position: r.position ?? r.Position ?? 0,
          title: r.title ?? r.Title ?? r.video_title ?? r.name ?? '',
          author_name: r.author_name ?? r.AuthorName ?? r.username ?? r.user ?? r.city ?? '',
          votes: r.votes ?? r.VoteCount ?? r.vote_count ?? 0,
          video_id: r.video_id ?? r.VideoID ?? r.VideoId ?? null,
        }));
      },
      error: (err) => {
        this.loading = false;
        this.error = err?.error?.message || err?.message || 'Error cargando rankings.';
      }
    });
  }

  getPositionClass(position: number): string {
    if (position === 1) return 'top-1';
    if (position === 2) return 'top-2';
    if (position === 3) return 'top-3';
    return '';
  }

  getTrophyIcon(position: number): string {
    if (position === 1) return 'emoji_events';
    if (position === 2) return 'military_tech';
    if (position === 3) return 'military_tech';
    return '';
  }

  getRowClass(position: number): string {
    return position <= 3 ? 'top-position' : '';
  }
}
