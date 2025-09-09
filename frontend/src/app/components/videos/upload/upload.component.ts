import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatTooltipModule } from '@angular/material/tooltip';

import { VideoService } from '../../../services/video.service';
import { AuthService } from '../../../services/auth.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-upload-video',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ReactiveFormsModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatSnackBarModule,
    MatCardModule,
    MatProgressBarModule,
    MatTooltipModule
  ],
  templateUrl: './upload.component.html',
  styleUrls: ['./upload.component.scss']
})
export class UploadVideoComponent {
  file?: File;
  title = '';
  uploading = false;
  dragOver = false;
  uploadProgress = 0;

  constructor(
    private fb: FormBuilder,
    private videoService: VideoService,
    private snackBar: MatSnackBar,
    private router: Router,
    private authService: AuthService
  ) { }

  onDragOver(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.dragOver = true;
  }

  onDragLeave(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.dragOver = false;
  }

  onDrop(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();
    this.dragOver = false;
    
    const files = event.dataTransfer?.files;
    if (files && files.length > 0) {
      this.validateAndSetFile(files[0]);
    }
  }

  onFileChange(event: any) {
    const files = event.target.files;
    if (files && files.length > 0) {
      this.validateAndSetFile(files[0]);
    }
  }

  validateAndSetFile(file: File) {
    // Validate type
    if (file.type !== 'video/mp4' && !file.name.toLowerCase().endsWith('.mp4')) {
      this.snackBar.open('Solo se permiten archivos MP4.', 'Cerrar', { duration: 3000 });
      return;
    }
    // Validate size <= 100MB
    const max = 100 * 1024 * 1024;
    if (file.size > max) {
      this.snackBar.open('El archivo supera el tamaño máximo de 100MB.', 'Cerrar', { duration: 3000 });
      return;
    }
    this.file = file;
  }

  removeFile() {
    this.file = undefined;
    this.uploadProgress = 0;
  }

  clearForm() {
    this.title = '';
    this.file = undefined;
    this.uploadProgress = 0;
  }

  getFileSizeString(): string {
    if (!this.file) return '';
    const size = this.file.size;
    if (size < 1024) return `${size} B`;
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  }

  submit() {
    if (!this.file) { this.snackBar.open('Seleccione un archivo MP4.', 'Cerrar', { duration: 2000 }); return; }
    if (!this.title || this.title.trim() === '') { this.snackBar.open('Ingrese un título.', 'Cerrar', { duration: 2000 }); return; }
    // ensure user is logged in
    const token = this.authService.getToken();
    if (!token) { this.snackBar.open('Debe iniciar sesión para subir videos.', 'Cerrar', { duration: 2000 }); return; }

    const fd = new FormData();
    fd.append('title', this.title.trim());
    fd.append('video', this.file as Blob, this.file!.name);

    this.uploading = true;
    this.videoService.uploadVideo(fd).subscribe({
      next: (res) => {
        this.uploading = false;
        const message = res?.message || 'Video subido correctamente.';
        const taskId = res?.task_id;
        this.snackBar.open(`${message}` + (taskId ? ` (task: ${taskId})` : ''), 'Cerrar', { duration: 5000 });
        // Optionally navigate to videos list
        this.router.navigate(['/dashboard']);
      },
      error: (err) => {
        this.uploading = false;
        const msg = err?.error?.message || err?.message || 'Error al subir el video.';
        this.snackBar.open(msg, 'Cerrar', { duration: 5000 });
      }
    });
  }
}
