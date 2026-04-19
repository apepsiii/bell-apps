import os
import sys
import json
import base64
import numpy as np
from io import BytesIO
from datetime import datetime
from contextlib import asynccontextmanager

import cv2
from fastapi import FastAPI, File, UploadFile, HTTPException, BackgroundTasks
from fastapi.responses import JSONResponse
from pydantic import BaseModel
import face_recognition
from PIL import Image
import mysql.connector
from mysql.connector import Error as MySQLError
from dotenv import load_dotenv

load_dotenv()

# Database configuration
DB_CONFIG = {
    "host": os.getenv("DB_HOST", "localhost"),
    "port": int(os.getenv("DB_PORT", "3306")),
    "user": os.getenv("DB_USER", "root"),
    "password": os.getenv("DB_PASSWORD", ""),
    "database": os.getenv("DB_NAME", "bell"),
}


def get_db_connection():
    """Get MySQL database connection"""
    try:
        conn = mysql.connector.connect(**DB_CONFIG)
        return conn
    except MySQLError as e:
        print(f"Database connection error: {e}")
        return None


def image_to_encoding(image_bytes: bytes) -> list:
    """Convert image bytes to 128-dimensional face encoding"""
    pil_image = Image.open(BytesIO(image_bytes))
    pil_image = pil_image.convert("RGB")
    opencv_image = np.array(pil_image)
    opencv_image = cv2.cvtColor(opencv_image, cv2.COLOR_RGB2BGR)

    face_encodings = face_recognition.face_encodings(opencv_image)

    if len(face_encodings) == 0:
        return None

    if len(face_encodings) > 1:
        return None

    return face_encodings[0].tolist()


def save_face_encoding(student_id: int, encoding: list, image_path: str = None) -> bool:
    """Save face encoding to database"""
    conn = get_db_connection()
    if not conn:
        return False

    try:
        cursor = conn.cursor()
        encoding_json = json.dumps(encoding)

        cursor.execute(
            """
            INSERT INTO student_faces (student_id, encoding, image_path, registered_at)
            VALUES (%s, %s, %s, NOW())
            ON DUPLICATE KEY UPDATE
            encoding = VALUES(encoding),
            image_path = VALUES(image_path),
            registered_at = NOW()
        """,
            (student_id, encoding_json, image_path),
        )

        conn.commit()
        cursor.close()
        conn.close()
        return True
    except MySQLError as e:
        print(f"Error saving face encoding: {e}")
        return False


def get_all_encodings():
    """Get all student face encodings from database"""
    conn = get_db_connection()
    if not conn:
        return []

    try:
        cursor = conn.cursor(dictionary=True)
        cursor.execute("""
            SELECT sf.student_id, sf.encoding, s.name, s.class_id, c.name as class_name
            FROM student_faces sf
            JOIN students s ON sf.student_id = s.id
            LEFT JOIN classes c ON s.class_id = c.id
        """)

        results = cursor.fetchall()
        cursor.close()
        conn.close()

        encodings = []
        for row in results:
            encoding_list = json.loads(row["encoding"])
            encodings.append(
                {
                    "student_id": row["student_id"],
                    "name": row["name"],
                    "class_name": row["class_name"] or "",
                    "encoding": encoding_list,
                }
            )

        return encodings
    except MySQLError as e:
        print(f"Error fetching encodings: {e}")
        return []


def recognize_face(unknown_encoding: list, threshold: float = 0.45) -> dict:
    """Compare unknown face with all known encodings"""
    known_encodings = get_all_encodings()

    if not known_encodings:
        return {"matched": False, "message": "No registered faces"}

    unknown_np = np.array(unknown_encoding)

    best_match = None
    best_distance = float("inf")

    for known in known_encodings:
        known_np = np.array(known["encoding"])
        distance = face_recognition.face_distance([known_np], unknown_np)[0]

        if distance < best_distance:
            best_distance = distance
            best_match = known

    if best_distance <= threshold:
        return {
            "matched": True,
            "student_id": best_match["student_id"],
            "name": best_match["name"],
            "class_name": best_match["class_name"],
            "distance": float(best_distance),
        }

    return {
        "matched": False,
        "distance": float(best_distance),
        "message": "Face not recognized",
    }


@asynccontextmanager
async def lifespan(app: FastAPI):
    print("Face Recognition Service Starting...")
    yield
    print("Face Recognition Service Shutting Down...")


app = FastAPI(
    title="SmartBell Face Recognition Service",
    description="Microservice for face encoding and verification",
    version="1.0.0",
    lifespan=lifespan,
)


class RegisterRequest(BaseModel):
    student_id: int
    name: str
    class_name: str = ""


class VerifyRequest(BaseModel):
    image_base64: str


class HealthResponse(BaseModel):
    status: str
    database: str
    face_count: int


@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint"""
    conn = get_db_connection()
    db_status = "connected" if conn else "disconnected"
    if conn:
        conn.close()

    face_count = len(get_all_encodings())

    return HealthResponse(status="healthy", database=db_status, face_count=face_count)


@app.post("/encode")
async def encode_face(
    student_id: int, name: str, class_name: str = "", image: UploadFile = File(...)
):
    """Register a student's face from uploaded image"""
    try:
        image_bytes = await image.read()

        encoding = image_to_encoding(image_bytes)

        if encoding is None:
            return JSONResponse(
                status_code=400,
                content={
                    "error": "No face detected or multiple faces found. Please use a clear, single face image."
                },
            )

        pil_image = Image.open(BytesIO(image_bytes))
        image_path = f"faces/{student_id}_{datetime.now().strftime('%Y%m%d%H%M%S')}.jpg"
        os.makedirs("faces", exist_ok=True)
        pil_image.save(image_path)

        success = save_face_encoding(student_id, encoding, image_path)

        if success:
            return {
                "success": True,
                "student_id": student_id,
                "message": f"Face registered for {name}",
            }
        else:
            return JSONResponse(
                status_code=500,
                content={"error": "Failed to save face encoding to database"},
            )

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.post("/verify")
async def verify_face(request: VerifyRequest):
    """Verify a face against registered faces"""
    try:
        image_data = request.image_base64

        if "," in image_data:
            image_data = image_data.split(",")[1]

        image_bytes = base64.b64decode(image_data)

        encoding = image_to_encoding(image_bytes)

        if encoding is None:
            return JSONResponse(
                status_code=400,
                content={"error": "No face detected or multiple faces found"},
            )

        result = recognize_face(encoding)

        return result

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.post("/verify_file")
async def verify_face_file(image: UploadFile = File(...)):
    """Verify a face from uploaded file"""
    try:
        image_bytes = await image.read()

        encoding = image_to_encoding(image_bytes)

        if encoding is None:
            return JSONResponse(
                status_code=400,
                content={"error": "No face detected or multiple faces found"},
            )

        result = recognize_face(encoding)

        return result

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.delete("/face/{student_id}")
async def delete_face(student_id: int):
    """Delete a student's face encoding"""
    conn = get_db_connection()
    if not conn:
        return JSONResponse(
            status_code=500, content={"error": "Database connection failed"}
        )

    try:
        cursor = conn.cursor()
        cursor.execute("DELETE FROM student_faces WHERE student_id = %s", (student_id,))
        conn.commit()
        cursor.close()
        conn.close()

        return {"success": True, "message": f"Face deleted for student {student_id}"}
    except MySQLError as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.get("/faces")
async def list_faces():
    """List all registered faces"""
    encodings = get_all_encodings()
    return {
        "count": len(encodings),
        "faces": [
            {
                "student_id": e["student_id"],
                "name": e["name"],
                "class_name": e["class_name"],
            }
            for e in encodings
        ],
    }


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8001"))
    uvicorn.run(app, host="0.0.0.0", port=port)
