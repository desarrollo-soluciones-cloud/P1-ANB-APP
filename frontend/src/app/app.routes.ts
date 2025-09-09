import { Routes } from '@angular/router';
import { AuthGuard } from './guards/auth.guard';

export const routes: Routes = [
  {
    path: 'auth',
    children: [
      { 
        path: 'login', 
        loadComponent: () => import('./components/auth/login/login.component').then(c => c.LoginComponent)
      },
      { 
        path: 'register', 
        loadComponent: () => import('./components/auth/register/register.component').then(c => c.RegisterComponent)
      },
      { 
        path: '', 
        redirectTo: 'login', 
        pathMatch: 'full' 
      }
    ]
  },
  { 
    path: 'dashboard', 
    loadComponent: () => import('./components/dashboard/dashboard.component').then(c => c.DashboardComponent),
    canActivate: [AuthGuard],
    children: [
      {
        path: 'videos/upload',
        loadComponent: () => import('./components/videos/upload/upload.component').then(c => c.UploadVideoComponent)
      },
      {
        path: 'videos',
        loadComponent: () => import('./components/videos/list/list.component').then(c => c.ListVideosComponent)
      },
      {
        path: 'rankings',
        loadComponent: () => import('./components/videos/ranking/ranking.component').then(c => c.RankingComponent)
      },
      {
        path: 'public/videos',
        loadComponent: () => import('./components/videos/public/public.component').then(c => c.PublicVideosComponent)
      },
      {
        path: 'videos/:id',
        loadComponent: () => import('./components/videos/detail/detail.component').then(c => c.VideoDetailComponent)
      },
      {
        path: '',
        redirectTo: 'videos',
        pathMatch: 'full'
      }
    ]
  },
  { 
    path: '', 
    redirectTo: '/dashboard', 
    pathMatch: 'full' 
  },
  {
    path: 'public',
    children: [
      {
        path: 'videos',
        loadComponent: () => import('./components/videos/public/public.component').then(c => c.PublicVideosComponent)
      },
      { path: '', redirectTo: 'videos', pathMatch: 'full' }
    ]
  },
  { 
    path: '**', 
    redirectTo: '/auth/login' 
  }
];
