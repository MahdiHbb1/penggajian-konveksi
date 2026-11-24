-- Pemilik: Mahdi | Modul 1 | Aturan: 3
-- Seed hanya berisi dua shift awal. Tidak ada akun atau password.

USE penggajian_konveksi;

INSERT INTO shift (nama, jam_mulai, jam_selesai) VALUES
  ('Pagi', '07:00:00', '15:00:00'),
  ('Siang', '15:00:00', '23:00:00');

-- Pemilik: Fadli | Modul 6 | Aturan: 12
-- Akun pengguna harus dibuat secara manual oleh administrator menggunakan tool:
--   go run internal/auth/genhash/main.go <password>
-- Kemudian jalankan SQL:
--   INSERT INTO pengguna (username, password_hash, role) VALUES ('<username>', '<hash>', 'ADMIN');
-- Jangan menyimpan password default dalam source code atau seed.
