import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterModule, RouterOutlet } from '@angular/router';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatListModule } from '@angular/material/list';
import { MatButtonModule } from '@angular/material/button';
import { MatSnackBarModule, MatSnackBar } from '@angular/material/snack-bar';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatMenuModule } from '@angular/material/menu';
import { MatCardModule } from '@angular/material/card';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    RouterOutlet,
    MatButtonModule,
    MatIconModule,
    MatSidenavModule,
    MatListModule,
    MatSnackBarModule,
    MatToolbarModule,
    MatMenuModule,
    MatCardModule,
    MatDividerModule,
    MatTooltipModule
  ],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  currentUser: any = null;
  navigationItems = [
    {
      label: 'Mis Videos',
      icon: 'video_library',
      route: '/dashboard/videos',
      description: 'Gestiona tus videos subidos'
    },
    {
      label: 'Subir Video',
      icon: 'cloud_upload',
      route: '/dashboard/videos/upload',
      description: 'Sube un nuevo video'
    },
    {
      label: 'Videos Públicos',
      icon: 'public',
      route: '/dashboard/public/videos',
      description: 'Vota por videos de la comunidad'
    },
    {
      label: 'Rankings',
      icon: 'leaderboard',
      route: '/dashboard/rankings',
      description: 'Ve la tabla de clasificación'
    }
  ];

  constructor(
    private authService: AuthService,
    private router: Router,
    private snackBar: MatSnackBar
  ) {}

  ngOnInit() {
    // Check authentication first
    const token = localStorage.getItem('token');
    if (!token) {
      console.log('No authentication token found, redirecting to login');
      this.router.navigate(['/auth/login']);
      return;
    }

    this.currentUser = this.authService.getCurrentUser();
    if (!this.currentUser) {
      console.log('No current user found, redirecting to login');
      this.router.navigate(['/auth/login']);
      return;
    }

    // Navigate to Videos by default
    if (this.router.url === '/dashboard' || this.router.url === '/dashboard/') {
      this.router.navigate(['/dashboard/videos']);
    }
  }

  logout() {
    this.authService.logout();
    this.router.navigate(['/auth/login']);
  }

  getUserInitials(): string {
    const user = this.currentUser;
    if (!user) return 'U';
    
    const name = user.username || user.name || user.email;
    if (!name) return 'U';
    
    return name.charAt(0).toUpperCase();
  }

  getUserDisplayName(): string {
    const user = this.currentUser;
    return user?.username || user?.name || 'Usuario';
  }
}
