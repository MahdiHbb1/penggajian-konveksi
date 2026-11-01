-- Pemilik: Mahdi | Modul 1 | Aturan: 3
-- Seed hanya berisi dua shift awal. Tidak ada akun atau password.

USE penggajian_konveksi;

INSERT INTO shift (nama, jam_mulai, jam_selesai) VALUES
  ('Pagi', '07:00:00', '15:00:00'),
  ('Siang', '15:00:00', '23:00:00');
