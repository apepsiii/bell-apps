import os
import json
import base64
import numpy as np
from io import BytesIO
from datetime import datetime
from contextlib import asynccontextmanager

import cv2
from fastapi import FastAPI, File, UploadFile
from fastapi.responses import JSONResponse
from PIL import Image
import mysql.connector
from mysql.connector import Error as MySQLError
from dotenv import load_dotenv

load_dotenv()

DB_CONFIG = {
    "host": os.getenv("DB_HOST", "localhost"),
    "port": int(os.getenv("DB_PORT", "3306")),
    "user": os.getenv("DB_USER", "root"),
    "password": os.getenv("DB_PASSWORD", ""),
    "database": os.getenv("DB_NAME", "bell"),
}

FACE_CASCADE = cv2.CascadeClassifier(
    cv2.data.haarcascades + "haarcascade_frontalface_default.xml"
)


def get_db_connection():
    try:
        conn = mysql.connector.connect(**DB_CONFIG)
        return conn
    except MySQLError as e:
        print(f"Database connection error: {e}")
        return None


def detect_and_crop_face(image):
    """Detect face in image and return cropped face"""
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    faces = FACE_CASCADE.detectMultiScale(gray, 1.3, 5)

    if len(faces) == 0:
        return None

    if len(faces) > 1:
        return None

    x, y, w, h = faces[0]
    face = image[y : y + h, x : x + w]
    return cv2.resize(face, (150, 150))


def image_to_encoding(image_bytes):
    """Convert image bytes to face encoding using LBPH"""
    pil_image = Image.open(BytesIO(image_bytes))
    pil_image = pil_image.convert("RGB")
    opencv_image = np.array(pil_image)
    opencv_image = cv2.cvtColor(opencv_image, cv2.COLOR_RGB2BGR)

    face = detect_and_crop_face(opencv_image)

    if face is None:
        return None

    gray = cv2.cvtColor(face, cv2.COLOR_BGR2GRAY)

    recognizer = cv2.face.LBPHFaceRecognizer_create()
    recognizer.train([gray], np.array([0]))

    return {
        "face": face.tolist() if isinstance(face, np.ndarray) else face,
        "gray": gray.tolist() if isinstance(gray, np.ndarray) else gray,
    }


def save_face_encoding(student_id, encoding, image_path=None):
    conn = get_db_connection()
    if not conn:
        return False

    try:
        cursor = conn.cursor()
        face_json = json.dumps(encoding)

        cursor.execute(
            """
            INSERT INTO student_faces (student_id, encoding, image_path, registered_at)
            VALUES (%s, %s, %s, NOW())
            ON DUPLICATE KEY UPDATE
            encoding = VALUES(encoding),
            image_path = VALUES(image_path),
            registered_at = NOW()
        """,
            (student_id, face_json, image_path),
        )

        conn.commit()
        cursor.close()
        conn.close()
        return True
    except MySQLError as e:
        print(f"Error saving face encoding: {e}")
        return False


def get_all_encodings():
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
            encoding_data = json.loads(row["encoding"])
            encodings.append(
                {
                    "student_id": row["student_id"],
                    "name": row["name"],
                    "class_name": row["class_name"] or "",
                    "gray": encoding_data.get("gray", []),
                }
            )

        return encodings
    except MySQLError as e:
        print(f"Error fetching encodings: {e}")
        return []


def recognize_face(gray_face, threshold=80):
    known_encodings = get_all_encodings()

    if not known_encodings:
        return {"matched": False, "message": "No registered faces"}

    recognizer = cv2.face.LBPHFaceRecognizer_create()

    best_match = None
    best_distance = float("inf")

    for known in known_encodings:
        try:
            known_gray = np.array(known.get("gray", []), dtype=np.uint8)
            if len(known_gray) == 0:
                continue

            known_ids = np.array([0])
            recognizer.setTrainLabels([0])
            recognizer.train([gray_face], known_ids)

            label, distance = recognizer.predict(gray_face)

            if distance < best_distance:
                best_distance = distance
                best_match = known
        except Exception as e:
            print(f"Error comparing face: {e}")
            continue

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
async def lifespan(app):
    print("Face Recognition Service Starting...")
    yield
    print("Face Recognition Service Shutting Down...")


app = FastAPI(title="SmartBell Face Recognition", version="1.0.0", lifespan=lifespan)


@app.get("/health")
async def health_check():
    conn = get_db_connection()
    db_status = "connected" if conn else "disconnected"
    if conn:
        conn.close()

    face_count = len(get_all_encodings())

    return {"status": "healthy", "database": db_status, "face_count": face_count}


@app.post("/encode")
async def encode_face(
    student_id: int, name: str, class_name: str = "", image: UploadFile = File(...)
):
    try:
        image_bytes = await image.read()

        pil_image = Image.open(BytesIO(image_bytes))
        pil_image = pil_image.convert("RGB")
        opencv_image = np.array(pil_image)
        opencv_image = cv2.cvtColor(opencv_image, cv2.COLOR_RGB2BGR)

        face = detect_and_crop_face(opencv_image)

        if face is None:
            return JSONResponse(
                status_code=400,
                content={"error": "No face detected or multiple faces found"},
            )

        gray = cv2.cvtColor(face, cv2.COLOR_BGR2GRAY)

        os.makedirs("faces", exist_ok=True)
        image_path = f"faces/{student_id}_{datetime.now().strftime('%Y%m%d%H%M%S')}.jpg"
        cv2.imwrite(image_path, face)

        encoding = {"face": face.tolist(), "gray": gray.tolist()}
        success = save_face_encoding(student_id, encoding, image_path)

        if success:
            return {
                "success": True,
                "student_id": student_id,
                "message": f"Face registered for {name}",
            }
        else:
            return JSONResponse(
                status_code=500, content={"error": "Failed to save face encoding"}
            )

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.post("/verify")
async def verify_face(request: dict):
    try:
        image_data = request.get("image_base64", "")

        if "," in image_data:
            image_data = image_data.split(",")[1]

        image_bytes = base64.b64decode(image_data)

        pil_image = Image.open(BytesIO(image_bytes))
        pil_image = pil_image.convert("RGB")
        opencv_image = np.array(pil_image)
        opencv_image = cv2.cvtColor(opencv_image, cv2.COLOR_RGB2BGR)

        face = detect_and_crop_face(opencv_image)

        if face is None:
            return JSONResponse(status_code=400, content={"error": "No face detected"})

        gray = cv2.cvtColor(face, cv2.COLOR_BGR2GRAY)

        result = recognize_face(gray)

        return result

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.post("/verify_file")
async def verify_face_file(image: UploadFile = File(...)):
    try:
        image_bytes = await image.read()

        pil_image = Image.open(BytesIO(image_bytes))
        pil_image = pil_image.convert("RGB")
        opencv_image = np.array(pil_image)
        opencv_image = cv2.cvtColor(opencv_image, cv2.COLOR_RGB2BGR)

        face = detect_and_crop_face(opencv_image)

        if face is None:
            return JSONResponse(status_code=400, content={"error": "No face detected"})

        gray = cv2.cvtColor(face, cv2.COLOR_BGR2GRAY)

        result = recognize_face(gray)

        return result

    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.delete("/face/{student_id}")
async def delete_face(student_id: int):
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
    encodings = get_all_encodings()
    return {"count": len(encodings), "faces": encodings}


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8001"))
    uvicorn.run(app, host="0.0.0.0", port=port)
