-- Pemilik: Mahdi | Modul 1 | Aturan: 3
-- Seed hanya berisi dua shift awal. Tidak ada akun atau password.

USE penggajian_konveksi;

INSERT INTO shift (nama, jam_mulai, jam_selesai) VALUES
  ('Pagi', '07:00:00', '15:00:00'),
  ('Siang', '15:00:00', '23:00:00');

-- Pemilik: Fadli | Modul 6 | Aturan: 12
-- Akun demo untuk pengembangan. Ganti password sebelum produksi.
-- Format hash: sha512:{salt}:{hash}
-- Password demo: lihat README atau tanyakan PIC Modul 6.
INSERT INTO pengguna (username, password_hash, role) VALUES
  ('admin',   'sha512:f079a1a334ac7193:80f9b11ad9467365fd96123748e8f867a4aee7dd1ab5d1bb9d9b636f0c044fc55176777605bb1f82c46ed0a85f1a8fc68470944d626841fe8c63fe1638dd9cfd', 'ADMIN'),
  ('pemilik', 'sha512:122816772f39ab4b:4face04908d1a941914d1565956cea3436ee45cb8fd874e41bd9be683844a73c8792f96e1733c6bb4e00f8265c502b2a28a80be07d9037de2e8158a1e908994d', 'PEMILIK');
