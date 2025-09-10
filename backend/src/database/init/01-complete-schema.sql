-- /backend/src/database/init/01-complete-schema.sql
-- Script completo que crea el esquema y pobla los datos automáticamente

-- =============================================================================
-- CREAR ESQUEMA DE BASE DE DATOS
-- =============================================================================

-- Crear tabla de usuarios
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    city VARCHAR(255),
    country VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Crear tabla de videos
CREATE TABLE IF NOT EXISTS videos (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'uploaded',
    original_url VARCHAR(500),
    processed_url VARCHAR(500),
    vote_count INTEGER DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Crear tabla de votos
CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    video_id INTEGER NOT NULL,
    voted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE CASCADE,
    UNIQUE(user_id, video_id)
);

-- Crear índices para mejorar rendimiento
CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_vote_count ON videos(vote_count DESC);
CREATE INDEX IF NOT EXISTS idx_votes_user_id ON votes(user_id);
CREATE INDEX IF NOT EXISTS idx_votes_video_id ON votes(video_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- =============================================================================
-- POBLAR CON DATOS DE PRUEBA
-- =============================================================================

-- Insertar usuarios de prueba (password: "password" hasheado con bcrypt)
INSERT INTO users (first_name, last_name, email, password, city, country, created_at, updated_at) VALUES
('Carlos', 'López', 'carlos@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Bogotá', 'Colombia', NOW(), NOW()),
('María', 'García', 'maria@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Medellín', 'Colombia', NOW(), NOW()),
('Luis', 'Rodríguez', 'luis@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Cali', 'Colombia', NOW(), NOW()),
('Ana', 'Martínez', 'ana@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Barranquilla', 'Colombia', NOW(), NOW()),
('Miguel', 'Hernández', 'miguel@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Cartagena', 'Colombia', NOW(), NOW()),
('Sofía', 'Ramírez', 'sofia@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Bucaramanga', 'Colombia', NOW(), NOW()),
('Diego', 'Torres', 'diego@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Pereira', 'Colombia', NOW(), NOW()),
('Camila', 'Vásquez', 'camila@anb.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Manizales', 'Colombia', NOW(), NOW());

-- Insertar videos de prueba (TODOS USAN EL MISMO ARCHIVO dunk_example.mp4)
-- Algunos procesados, otros en estado uploaded
INSERT INTO videos (user_id, title, status, original_url, processed_url, vote_count, uploaded_at, processed_at) VALUES
-- Videos de Carlos (user_id: 1)
(1, 'Jugada Espectacular de Carlos - Triple desde media cancha', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 45, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
(1, 'Triple Decisivo en el último segundo', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 38, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(1, 'Secuencia de tiros libres perfectos', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 22, NOW() - INTERVAL '3 days', NULL),
(1, 'Entrenamiento matutino - Técnica de dribleo', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 0, NOW() - INTERVAL '1 day', NULL),

-- Videos de María (user_id: 2)
(2, 'Defensa Perfecta de María - Robo y contraataque', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 52, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
(2, 'Asistencia Increíble sin mirar', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 34, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(2, 'Técnica de dribleo avanzado entre conos', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 18, NOW() - INTERVAL '2 days', NULL),
(2, 'Bloqueo defensivo coordinado', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 0, NOW() - INTERVAL '6 hours', NULL),

-- Videos de Luis (user_id: 3)
(3, 'Dunk Espectacular de Luis - Mate con giro 360°', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 67, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),
(3, 'Robo y Contraataque Lightning', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 41, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(3, 'Bloqueo defensivo épico - Rechazo total', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 29, NOW() - INTERVAL '4 days', NULL),
(3, 'Salto vertical - Entrenamiento de potencia', 'failed', '/uploads/originals/dunk_example.mp4', '', 0, NOW() - INTERVAL '2 days', NULL),

-- Videos de Ana (user_id: 4)
(4, 'Tiro libre bajo presión - Secuencia de 10/10', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 25, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(4, 'Jugada colectiva perfecta - Asistencia de lujo', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 31, NOW() - INTERVAL '3 days', NULL),
(4, 'Técnica de pivoteo y finalizacion', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 0, NOW() - INTERVAL '1 day', NULL),

-- Videos de Miguel (user_id: 5)
(5, 'Salto vertical impresionante - 95cm de elevación', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 19, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(5, 'Combinación de habilidades - Dribleo y tiro', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 33, NOW() - INTERVAL '2 days', NULL),
(5, 'Entrenamiento de resistencia', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 0, NOW() - INTERVAL '12 hours', NULL),

-- Videos de Sofía (user_id: 6)
(6, 'Triple desde la esquina - Técnica perfecta', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 14, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(6, 'Jugada creativa en el poste bajo', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 27, NOW() - INTERVAL '1 day', NULL),

-- Videos de Diego (user_id: 7)
(7, 'Asistencia no-look perfecta', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 16, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(7, 'Control de tempo y manejo de balón', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 21, NOW() - INTERVAL '1 day', NULL),

-- Videos de Camila (user_id: 8)
(8, 'Secuencia de triples - 8 de 10 intentos', 'processed', '/uploads/originals/dunk_example.mp4', '/uploads/processed/dunk_example.mp4', 35, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(8, 'Movimiento sin balón y tiro en suspensión', 'uploaded', '/uploads/originals/dunk_example.mp4', '', 0, NOW() - INTERVAL '3 hours', NULL);

-- Insertar votos realistas entre usuarios (solo para videos processed)
INSERT INTO votes (user_id, video_id, voted_at, created_at) VALUES
-- Carlos vota por videos procesados de otros
(1, 5, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),   -- Video procesado de María
(1, 6, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),   -- Video procesado de María
(1, 9, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),   -- Video procesado de Luis (el más popular)
(1, 10, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Luis
(1, 13, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),  -- Video procesado de Ana
(1, 16, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Miguel
(1, 18, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),  -- Video procesado de Sofía
(1, 20, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),  -- Video procesado de Diego

-- María vota por videos procesados de otros
(2, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),   -- Video procesado de Carlos (muy popular)
(2, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),   -- Video procesado de Carlos
(2, 9, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),   -- Video procesado de Luis
(2, 10, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Luis
(2, 13, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),  -- Video procesado de Ana
(2, 16, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Miguel
(2, 18, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),  -- Video procesado de Sofía
(2, 20, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),  -- Video procesado de Diego
(2, 22, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),  -- Video procesado de Camila

-- Luis vota por videos procesados de otros  
(3, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),   -- Video procesado de Carlos
(3, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),   -- Video procesado de Carlos
(3, 5, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),   -- Video procesado de María
(3, 6, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),   -- Video procesado de María
(3, 13, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),  -- Video procesado de Ana
(3, 16, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Miguel
(3, 18, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),  -- Video procesado de Sofía
(3, 20, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),  -- Video procesado de Diego
(3, 22, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),  -- Video procesado de Camila

-- Ana vota por videos procesados de otros
(4, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),   -- Video procesado de Carlos
(4, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),   -- Video procesado de Carlos
(4, 5, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),   -- Video procesado de María
(4, 6, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),   -- Video procesado de María
(4, 9, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),   -- Video procesado de Luis
(4, 10, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Luis
(4, 16, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Miguel
(4, 18, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),  -- Video procesado de Sofía
(4, 20, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),  -- Video procesado de Diego
(4, 22, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),  -- Video procesado de Camila

-- Miguel vota por videos procesados de otros
(5, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),   -- Video procesado de Carlos
(5, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),   -- Video procesado de Carlos
(5, 5, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),   -- Video procesado de María
(5, 6, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),   -- Video procesado de María
(5, 9, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),   -- Video procesado de Luis
(5, 10, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Luis
(5, 13, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),  -- Video procesado de Ana
(5, 18, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),  -- Video procesado de Sofía
(5, 20, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),  -- Video procesado de Diego

-- Sofía vota por videos procesados de otros
(6, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),   -- Video procesado de Carlos
(6, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),   -- Video procesado de Carlos
(6, 5, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),   -- Video procesado de María
(6, 9, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),   -- Video procesado de Luis
(6, 10, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Luis
(6, 13, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),  -- Video procesado de Ana
(6, 16, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Miguel
(6, 20, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),  -- Video procesado de Diego
(6, 22, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),  -- Video procesado de Camila

-- Diego vota por videos procesados de otros
(7, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),   -- Video procesado de Carlos
(7, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),   -- Video procesado de Carlos
(7, 5, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),   -- Video procesado de María
(7, 6, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),   -- Video procesado de María
(7, 9, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),   -- Video procesado de Luis
(7, 13, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),  -- Video procesado de Ana
(7, 16, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Miguel
(7, 18, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),  -- Video procesado de Sofía
(7, 22, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),  -- Video procesado de Camila

-- Camila vota por videos procesados de otros
(8, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),   -- Video procesado de Carlos
(8, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),   -- Video procesado de Carlos
(8, 5, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),   -- Video procesado de María
(8, 6, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),   -- Video procesado de María
(8, 9, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),   -- Video procesado de Luis
(8, 10, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Luis
(8, 13, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),  -- Video procesado de Ana
(8, 16, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),  -- Video procesado de Miguel
(8, 18, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),  -- Video procesado de Sofía
(8, 20, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days');  -- Video procesado de Diego