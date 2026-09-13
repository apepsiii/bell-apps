Beres, data udah lengkap. Sekarang saya kasih design sistem poin ideal — nyambung langsung sama struktur DB kamu, bukan teori kosong.

MASALAH SISTEM SEKARANG (berdasarkan data real)

1. Ranking dikuasai absensi doang — Nuri 620 poin cuma dari datang tepat waktu. Lomba tingkat provinsi max 100 poin, sekali menang kalah sama 5 hari rajin dateng. Bobotnya nggak adil.
2. rule_id di student_points kosong semua — poin otomatis nggak nyambung ke kategori, jadi laporan per kategori mustahil dibuat.
3. Tabel student_achievement_points & student_violation_points nggak kepake (0 baris) — prestasi & pelanggaran nggak masuk sistem poin.
4. Urutan SP1 (25) → SP3 (76) → hukuman. Sistemnya lebih "menakutkan" daripada "mendukung".



DESIGN IDEAL: "3 REKA + 1 KONVERSI"

Analoginya kayak ujian sekolah: nilai akhir bukan dari satu jenis soal doang, tapi bobot per komponen. Begini strukturnya:

1. PIJAKAN (otomatis, bobot 30%)
- Bukan akumulasi poin mentah, tapi Indeks Kehadiran per bulan: 0–100
- Rumus: hadir tepat waktu = nilai penuh, terlambat = 50%, alpha = 0. Streak 10 hari +bonus 10 poin indeks (reward konsistensi, bukan volume)
- Datang dari sistem RFID yang udah jalan — nggak ada kerja manual
- Efek samping bagus: hadir 30 hari vs 29 hari bedanya kecil, jadi nggak ada yang "gaptek" gara-gara sakit seminggu

2. PRESTASI (manual, bobot 50%)
- Pake tabel achievement_rules (36 rule, 10 kategori) yang udah ada — ini jantungnya
- Paritas kategori: max 200 poin per kategori per semester (Akademik, Keagamaan, Sosial, Wirausaha, Ekskul). Jadi siswa yang bukan "anak ibadah" atau bukan "anak olahraga" tetap punya jalur menang sendiri
- Semua lewat pending_achievement_points dulu → guru approve → baru masuk. Udah ada tabelnya, tinggal diaktifkan
- APPROVAL_THRESHOLD = 50 yang udah ada dipertahankan: poin ≥50 wajib bukti (foto/sertifikat)

3. PELANGGARAN (bobot 20%, tapi sifatnya rehabilitatif)
- Pake eskalasi points_1/2/3 di violation_rules yang udah ada (pelanggaran sama makin sering = makin berat) — ini udah bagus
- Yang ditambah: jalur pemulihan. Pelanggaran bisa "dilunasi" lewat kerja positif (piket ekstra, mentoring adik kelas) dengan konversi 1:1
- SP1/SP2/SP3 tetap jadi notifikasi konseling, bukan vonis — SP3 memicu pembicaraan ortu, bukan langsung sanksi berat

4. KONVERSI — Skor Komposit per Semester

Skor = (0.30 × Indeks Kehadiran) + (0.50 × Poin Prestasi ternormalisasi) + (0.20 × 100 - Beban Pelanggaran)

- Reset tiap semester — siswa baru di tengah semester nggak permanently tertinggal (fair)
- Leaderboard pakai skor komposit; poin mentah tetap dipake buat katalog reward

TINGKATAN REWARD (balik hukuman jadi hadiah)

Skema kamu udah punya threshold bagus, tinggal diposisikan ulang:
- CERT (100) → Sertifikat "Siswa Berprestasi" — dicapai semester pertama
- REWARD (126) → Tukar poin dengan reward fisik (buku, merchandise sekolah)
- WALUYA (151) → Gelar tertinggi + privilege (misal: dispensasi ekskul pilihan, acara pentas)
- SP1–SP3 tetap ada tapi sebagai "early warning konseling" — arahnya bantu, bukan hukum mati

ANTI-GAMING (biar fair beneran)
- RATE_LIMIT_PER_DAY = 50 + MAX_POINTS_PER_TRANSACTION = 100 — pertahankan, udah ada
- point_audit_log (tabel udah ada, 0 baris) wajib diisi setiap approve guru — jejak lengkap siapa kasih apa
- Poin keagamaan "laporan mandiri" dibatasi kecil (5–15) + verifikasi periodik — ini titik paling rawan diborong