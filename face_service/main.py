import os
import json
import base64
import sqlite3
import numpy as np
from io import BytesIO
from datetime import datetime
from contextlib import asynccontextmanager

import cv2
from fastapi import FastAPI, File, UploadFile
from fastapi.responses import JSONResponse
from PIL import Image
from dotenv import load_dotenv

load_dotenv()

DB_PATH = os.getenv("DB_PATH", "./faces.db")

FACE_CASCADE = cv2.CascadeClassifier(
    cv2.data.haarcascades + "haarcascade_frontalface_default.xml"
)


def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


def init_db():
    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS student_faces (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            student_id INTEGER NOT NULL,
            name TEXT DEFAULT '',
            class_name TEXT DEFAULT '',
            encoding TEXT NOT NULL,
            image_path TEXT,
            registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(student_id)
        )
    """)
    conn.commit()
    conn.close()


def save_face_encoding(student_id, name, class_name, encoding, image_path=None):
    conn = get_db()
    try:
        cursor = conn.cursor()
        encoding_json = json.dumps(encoding)

        cursor.execute("""
            INSERT OR REPLACE INTO student_faces (student_id, name, class_name, encoding, image_path, registered_at)
            VALUES (?, ?, ?, ?, ?, datetime('now'))
        """, (student_id, name, class_name, encoding_json, image_path))

        conn.commit()
        conn.close()
        return True
    except Exception as e:
        print(f"Error saving face: {e}")
        conn.close()
        return False


def get_all_encodings():
    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("SELECT student_id, name, class_name, encoding FROM student_faces")

    rows = cursor.fetchall()
    conn.close()

    encodings = []
    for row in rows:
        try:
            encoding_data = json.loads(row["encoding"])
            encodings.append({
                "student_id": row["student_id"],
                "name": row["name"] or "Unknown",
                "class_name": row["class_name"] or "",
                "gray": encoding_data.get("gray", []),
            })
        except Exception as e:
            print(f"Error loading encoding for student {row['student_id']}: {e}")
            continue

    print(f"Loaded {len(encodings)} face encodings from database")
    return encodings


def detect_and_crop_face(image):
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    faces = FACE_CASCADE.detectMultiScale(gray, 1.3, 5)

    if len(faces) == 0:
        return None, None

    if len(faces) > 1:
        return None, None

    x, y, w, h = faces[0]
    face = image[y:y+h, x:x+w]
    face = cv2.resize(face, (150, 150))
    gray_face = cv2.cvtColor(face, cv2.COLOR_BGR2GRAY)

    return face, gray_face


def recognize_face(gray_face, threshold=80):
    known_encodings = get_all_encodings()

    if not known_encodings:
        return {"matched": False, "message": "No registered faces"}

    recognizer = cv2.face.LBPHFaceRecognizer_create()

    best_match = None
    best_distance = float("inf")

    for known in known_encodings:
        try:
            known_gray = known.get("gray", [])
            if not known_gray:
                continue

            known_img = cv2.resize(np.array(known_gray, dtype=np.uint8), (150, 150))
            known_img_rgb = cv2.cvtColor(known_img, cv2.COLOR_GRAY2BGR)

            label, distance = recognizer.predict(gray_face)

            if distance < best_distance:
                best_distance = distance
                best_match = known
        except Exception as e:
            print(f"Error comparing: {e}")
            continue

    print(f"Best match: {best_match['name'] if best_match else 'None'}, distance: {best_distance}")

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
        "message": f"Face not recognized (distance: {best_distance:.1f})",
    }


@asynccontextmanager
async def lifespan(app):
    print("Face Recognition Service Starting...")
    init_db()
    yield
    print("Face Recognition Service Shutting Down...")


app = FastAPI(title="SmartBell Face Recognition", version="2.0.0", lifespan=lifespan)


@app.get("/health")
async def health_check():
    init_db()
    face_count = len(get_all_encodings())
    return {"status": "healthy", "face_count": face_count}


@app.post("/register")
async def register_face(request: dict):
    try:
        student_id = request.get("student_id")
        name = request.get("name", "")
        class_name = request.get("class_name", "")
        image_base64 = request.get("image_base64", "")

        if not student_id:
            return JSONResponse(status_code=400, content={"error": "student_id is required"})

        if not image_base64:
            return JSONResponse(status_code=400, content={"error": "image_base64 is required"})

        if "," in image_base64:
            image_base64 = image_base64.split(",")[1]

        image_bytes = base64.b64decode(image_base64)

        pil_image = Image.open(BytesIO(image_bytes))
        pil_image = pil_image.convert("RGB")
        opencv_image = np.array(pil_image)
        opencv_image = cv2.cvtColor(opencv_image, cv2.COLOR_RGB2BGR)

        face, gray = detect_and_crop_face(opencv_image)

        if face is None:
            return JSONResponse(status_code=400, content={"error": "No face detected or multiple faces found"})

        os.makedirs("faces", exist_ok=True)
        image_path = f"faces/{student_id}_{datetime.now().strftime('%Y%m%d%H%M%S')}.jpg"
        cv2.imwrite(image_path, face)

        encoding = {"face": face.tolist(), "gray": [row.tolist() if hasattr(row, 'tolist') else row for row in gray.tolist()]}
        success = save_face_encoding(student_id, name, class_name, encoding, image_path)

        if success:
            return {
                "success": True,
                "student_id": student_id,
                "message": f"Face registered for {name}",
            }
        else:
            return JSONResponse(status_code=500, content={"error": "Failed to save face encoding"})

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.post("/verify")
async def verify_face(request: dict):
    try:
        image_data = request.get("image_base64", "")

        if not image_data:
            return JSONResponse(status_code=400, content={"error": "image_base64 is required"})

        if "," in image_data:
            image_data = image_data.split(",")[1]

        image_bytes = base64.b64decode(image_data)

        pil_image = Image.open(BytesIO(image_bytes))
        pil_image = pil_image.convert("RGB")
        opencv_image = np.array(pil_image)
        opencv_image = cv2.cvtColor(opencv_image, cv2.COLOR_RGB2BGR)

        face, gray = detect_and_crop_face(opencv_image)

        if face is None:
            return JSONResponse(status_code=400, content={"error": "No face detected"})

        gray_face = cv2.cvtColor(face, cv2.COLOR_BGR2GRAY)

        result = recognize_face(gray_face)

        return result

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.delete("/face/{student_id}")
async def delete_face(student_id: int):
    conn = get_db()
    cursor = conn.cursor()
    cursor.execute("DELETE FROM student_faces WHERE student_id = ?", (student_id,))
    conn.commit()
    conn.close()

    return {"success": True, "message": f"Face deleted for student {student_id}"}


@app.get("/faces")
async def list_faces():
    encodings = get_all_encodings()
    return {"count": len(encodings), "faces": encodings}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8001"))
    uvicorn.run(app, host="0.0.0.0", port=port)
