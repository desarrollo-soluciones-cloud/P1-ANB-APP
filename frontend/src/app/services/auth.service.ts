import { Injectable, Inject, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, BehaviorSubject, tap } from 'rxjs';
import { LoginRequest, CreateUserRequest, UserResponse, User } from '../models/user.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = environment.apiUrl;
  private currentUserSubject = new BehaviorSubject<User | null>(null);
  public currentUser$ = this.currentUserSubject.asObservable();

  constructor(
    private http: HttpClient,
    @Inject(PLATFORM_ID) private platformId: Object
  ) {
    // Check if user is logged in on service initialization
    this.checkAuthStatus();
  }

  private checkAuthStatus() {
    if (isPlatformBrowser(this.platformId)) {
      const token = localStorage.getItem('token');
      const userData = localStorage.getItem('currentUser');
      
      if (token && userData) {
        try {
          const user = JSON.parse(userData);
          this.currentUserSubject.next(user);
        } catch (error) {
          this.logout();
        }
      }
    }
  }

  // Adapted login to backend /auth/login which expects { email, password }
  login(credentials: LoginRequest): Observable<any> {
    // Support old form that uses 'username' by mapping to 'email'
    const payload: any = {
      email: (credentials as any).email || (credentials as any).username,
      password: (credentials as any).password
    };

  return this.http.post<any>(`${this.apiUrl}/api/v1/auth/login`, payload)
      .pipe(
        tap(response => {
          // Backend returns a TokenResponse with access_token
          const token = response?.access_token || response?.AccessToken || response?.accessToken || response?.token || response?.data?.token;
          if (token && isPlatformBrowser(this.platformId)) {
            localStorage.setItem('token', token);
            // store minimal current user info (email)
            localStorage.setItem('currentUser', JSON.stringify({ email: payload.email }));
            this.currentUserSubject.next({ email: payload.email } as any);
          }
        })
      );
  }

  // Register a new user via backend /auth/signup which expects JSON body
  register(userData: CreateUserRequest): Observable<any> {
    const payload = {
      first_name: userData.first_name,
      last_name: userData.last_name,
      email: userData.email,
      password: userData.password,
      password2: userData.password2,
      city: userData.city,
      country: userData.country
    };

  return this.http.post<any>(`${this.apiUrl}/api/v1/auth/signup`, payload);
  }

  logout() {
    if (isPlatformBrowser(this.platformId)) {
      localStorage.removeItem('token');
      localStorage.removeItem('currentUser');
    }
    this.currentUserSubject.next(null);
  }

  getToken(): string | null {
    if (isPlatformBrowser(this.platformId)) {
      return localStorage.getItem('token');
    }
    return null;
  }

  isLoggedIn(): boolean {
    return !!this.getToken();
  }

  getCurrentUser(): User | null {
    return this.currentUserSubject.value;
  }

  getAuthHeaders(): HttpHeaders {
    const token = this.getToken();
    return new HttpHeaders({
      'Authorization': token ? `Bearer ${token}` : '',
      'Content-Type': 'application/json'
    });
  }

  getAuthHeadersForFormData(): HttpHeaders {
    const token = this.getToken();
    return new HttpHeaders({
      'Authorization': token ? `Bearer ${token}` : ''
    });
  }
}
