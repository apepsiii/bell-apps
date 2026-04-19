# SmartBell Face Recognition Microservice

Microservice untuk face encoding dan verification. Berbagi database dengan Go backend (MySQL/MariaDB).

## Requirements

- Python 3.9+
- MySQL/MariaDB (bisa sharing dengan Go backend)
- Webcam untuk capture

## Installation

```bash
cd face_service
pip install -r requirements.txt
```

## Configuration

Buat file `.env` atau set environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=secret
export DB_NAME=bell
export PORT=8001
```

## Run

```bash
uvicorn main:app --reload --host 0.0.0.0 --port 8001
```

Atau langsung:
```bash
python main.py
```

## API Endpoints

### Health Check
```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "database": "connected",
  "face_count": 15
}
```

### Register Face (Encode)
```
POST /encode?student_id=1&name=Ahmad&class_name=X IPA 1
Content-Type: multipart/form-data
Body: image (file)
```

Response:
```json
{
  "success": true,
  "student_id": 1,
  "message": "Face registered for Ahmad"
}
```

### Verify Face (Base64)
```
POST /verify
Content-Type: application/json
Body: {"image_base64": "data:image/jpeg;base64,..."}
```

Response:
```json
{
  "matched": true,
  "student_id": 1,
  "name": "Ahmad",
  "class_name": "X IPA 1",
  "distance": 0.32
}
```

### Verify Face (File Upload)
```
POST /verify_file
Content-Type: multipart/form-data
Body: image (file)
```

### List All Faces
```
GET /faces
```

Response:
```json
{
  "count": 2,
  "faces": [
    {"student_id": 1, "name": "Ahmad", "class_name": "X IPA 1"},
    {"student_id": 2, "name": "Budi", "class_name": "X IPA 2"}
  ]
}
```

### Delete Face
```
DELETE /face/{student_id}
```

## Database Schema

Service ini membuat tabel `student_faces` sendiri:

```sql
CREATE TABLE IF NOT EXISTS student_faces (
    id INT PRIMARY KEY AUTO_INCREMENT,
    student_id INT NOT NULL,
    encoding MEDIUMTEXT NOT NULL,
    image_path VARCHAR(500),
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    UNIQUE KEY unique_student_face (student_id)
);
```

Note: Tabel `students` tetap ada di Go backend.

## Face Recognition Flow

1. **Register**: Admin upload foto siswa → `/encode` → face encoding disimpan ke DB
2. **Verify**: Camera capture → `/verify` → compare dengan semua encoding → return matched student

## Troubleshooting

### "No face detected"
- Foto terlalu gelap atau terlalu terang
- Wajah terlalu miring
- Ada lebih dari 1 wajah di foto

### "Face not recognized"
- Wajah belum terdaftar
- Foto terlalu berbeda dari yang terdaftar (angle, lighting)
- Distance threshold terlalu ketat (default 0.45)

### Database connection failed
- Pastikan MySQL running dan credentials benar
- Cek `SHOW DATABASES` untuk verify connection
