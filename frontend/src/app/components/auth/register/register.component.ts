import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService } from '../../../services/auth.service';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatSnackBarModule,
    MatProgressSpinnerModule
  ],
  templateUrl: './register.component.html',
  styleUrl: './register.component.scss'
})
export class RegisterComponent implements OnInit {
  registerForm: FormGroup;
  isLoading = false;
  hidePassword = true;
  hideConfirmPassword = true;
  selectedFile: File | null = null;
  imagePreview: string | null = null;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
    private snackBar: MatSnackBar
  ) {
    this.registerForm = this.fb.group({
      first_name: ['', [Validators.required]],
      last_name: ['', [Validators.required]],
      email: ['', [Validators.required, Validators.email]],
      password: ['', [Validators.required, Validators.minLength(6)]],
      password2: ['', [Validators.required]],
      city: ['', [Validators.required]],
      country: ['', [Validators.required]]
    }, {
      validators: this.passwordMatchValidator
    });
  }

  ngOnInit(): void {}

  // Custom validator to check if passwords match
  passwordMatchValidator(form: FormGroup) {
    const password = form.get('password');
    const confirmPassword = form.get('confirmPassword');
    
    if (password && confirmPassword && password.value !== confirmPassword.value) {
      confirmPassword.setErrors({ passwordMismatch: true });
      return { passwordMismatch: true };
    }
    
    return null;
  }

  onFileSelected(event: any): void {
  // Avatar upload removed for new registration flow; keep function for backward compatibility
  return;
  }

  removeImage(): void {
    this.selectedFile = null;
    this.imagePreview = null;
  }

  onSubmit(): void {
    if (this.registerForm.valid) {
      this.isLoading = true;

      const payload = {
        first_name: this.registerForm.get('first_name')?.value,
        last_name: this.registerForm.get('last_name')?.value,
        email: this.registerForm.get('email')?.value,
        password: this.registerForm.get('password')?.value,
        password2: this.registerForm.get('password2')?.value,
        city: this.registerForm.get('city')?.value,
        country: this.registerForm.get('country')?.value
      };

      this.authService.register(payload).subscribe({
        next: (response) => {
          this.isLoading = false;
          // On success backend returns created user; redirect to login
          this.snackBar.open('Registration successful! Please sign in.', 'Close', { duration: 3000 });
          this.router.navigate(['/auth/login']);
        },
        error: (error) => {
          this.isLoading = false;
          console.error('Registration error:', error);
          const errorMessage = error.error?.error || error.error?.message || error.message || 'Registration failed. Please try again.';
          this.snackBar.open(errorMessage, 'Close', { duration: 5000 });
        }
      });
    }
  }

  goToLogin(): void {
    this.router.navigate(['/auth/login']);
  }
}
