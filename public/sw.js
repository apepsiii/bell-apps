const CACHE_NAME = 'bel-sekolah-v1';
const ASSETS = [
  '/',
  '/index.html',
  '/manifest.json',
  '/api/sync', // Cache data jadwal terakhir
  '/audio/bell-masuk.mp3', // Daftarkan file audio utama
  '/audio/bell-istirahat.mp3'
];

// Install Service Worker & Cache Assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(ASSETS);
    })
  );
});

// Strategi: Network First, Fallback to Cache
// Artinya: Coba ambil data terbaru dari VPS, jika gagal (offline), ambil dari Cache.
self.addEventListener('fetch', (event) => {
  event.respondWith(
    fetch(event.request).catch(() => {
      return caches.match(event.request);
    })
  );
});