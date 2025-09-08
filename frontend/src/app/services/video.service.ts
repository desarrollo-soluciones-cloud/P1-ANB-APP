import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { AuthService } from './auth.service';

@Injectable({
  providedIn: 'root'
})
export class VideoService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient, private authService: AuthService) { }

  uploadVideo(formData: FormData): Observable<any> {
    const headers = this.authService.getAuthHeadersForFormData();
    return this.http.post<any>(`${this.apiUrl}/api/v1/videos/upload`, formData, { headers });
  }

  listMyVideos(): Observable<any> {
    const headers = this.authService.getAuthHeaders();
    return this.http.get<any>(`${this.apiUrl}/api/v1/videos`, { headers });
  }

  listPublicVideos(): Observable<any> {
    // Public endpoint, no auth required
    return this.http.get<any>(`${this.apiUrl}/api/v1/public/videos`);
  }

  getVideoById(videoId: string | number): Observable<any> {
    const headers = this.authService.getAuthHeaders();
    return this.http.get<any>(`${this.apiUrl}/api/v1/videos/${videoId}`, { headers });
  }

  deleteVideo(videoId: string | number): Observable<any> {
    const headers = this.authService.getAuthHeaders();
    return this.http.delete<any>(`${this.apiUrl}/api/v1/videos/${videoId}`, { headers });
  }

  // Vote for a public video. Requires authentication. POST /api/v1/public/videos/:video_id/vote
  voteVideo(videoId: string | number): Observable<any> {
    const headers = this.authService.getAuthHeaders();
    return this.http.post<any>(`${this.apiUrl}/api/v1/public/videos/${videoId}/vote`, {}, { headers });
  }

  // Remove vote for a public video. Requires authentication. DELETE /api/v1/public/videos/:video_id/vote
  unvoteVideo(videoId: string | number): Observable<any> {
    const headers = this.authService.getAuthHeaders();
    return this.http.delete<any>(`${this.apiUrl}/api/v1/public/videos/${videoId}/vote`, { headers });
  }

  // Get rankings (public)
  listRankings(params?: any): Observable<any> {
    // params can contain pagination & filters (e.g., city)
    const options: any = { params: params || {} };
    return this.http.get<any>(`${this.apiUrl}/api/v1/public/rankings`, options);
  }

  // Get user information by user ID (try common endpoint patterns)
  getUserById(userId: string | number): Observable<any> {
    // Try public endpoint first (common pattern)
    return this.http.get<any>(`${this.apiUrl}/api/v1/users/${userId}`);
  }

  // Alternative endpoint pattern for getting user info
  getUserProfile(userId: string | number): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/api/v1/public/users/${userId}`);
  }
}
