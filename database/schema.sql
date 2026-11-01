-- Pemilik: Mahdi | Modul 1 | Aturan: 1, 2, 3, 8, 12, 13, 15
-- Skema MySQL 8 untuk Sistem Penggajian Konveksi Mitra Jaya.
-- Data tidak dihapus permanen. Semua foreign key memakai RESTRICT.

CREATE DATABASE IF NOT EXISTS penggajian_konveksi
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;

USE penggajian_konveksi;

CREATE TABLE pekerja (
  id BIGINT NOT NULL AUTO_INCREMENT,
  kode VARCHAR(20) NOT NULL,
  nama VARCHAR(100) NOT NULL,
  jenis VARCHAR(10) NOT NULL,
  tanggal_mulai DATE NOT NULL,
  tanggal_berhenti DATE NULL,
  aktif BOOLEAN NOT NULL DEFAULT TRUE,
  PRIMARY KEY (id),
  UNIQUE KEY uq_pekerja_kode (kode),
  KEY idx_pekerja_nama (nama),
  KEY idx_pekerja_aktif (aktif),
  CONSTRAINT chk_pekerja_jenis CHECK (jenis IN ('HARIAN', 'BULANAN')),
  CONSTRAINT chk_pekerja_tanggal CHECK (tanggal_berhenti IS NULL OR tanggal_berhenti >= tanggal_mulai)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE riwayat_upah (
  id BIGINT NOT NULL AUTO_INCREMENT,
  pekerja_id BIGINT NOT NULL,
  nominal BIGINT NOT NULL,
  berlaku_mulai DATE NOT NULL,
  alasan VARCHAR(100) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_upah_pekerja_tanggal (pekerja_id, berlaku_mulai),
  KEY idx_upah_pekerja (pekerja_id),
  CONSTRAINT chk_upah_nominal CHECK (nominal > 0),
  CONSTRAINT fk_upah_pekerja FOREIGN KEY (pekerja_id) REFERENCES pekerja (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE shift (
  id BIGINT NOT NULL AUTO_INCREMENT,
  nama VARCHAR(30) NOT NULL,
  jam_mulai TIME NOT NULL,
  jam_selesai TIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_shift_nama (nama),
  CONSTRAINT chk_shift_jam CHECK (jam_selesai > jam_mulai)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE jadwal_shift (
  id BIGINT NOT NULL AUTO_INCREMENT,
  pekerja_id BIGINT NOT NULL,
  shift_id BIGINT NOT NULL,
  tanggal DATE NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_jadwal_pekerja_tanggal (pekerja_id, tanggal),
  KEY idx_jadwal_tanggal (tanggal),
  KEY idx_jadwal_shift (shift_id),
  CONSTRAINT fk_jadwal_pekerja FOREIGN KEY (pekerja_id) REFERENCES pekerja (id) ON DELETE RESTRICT,
  CONSTRAINT fk_jadwal_shift FOREIGN KEY (shift_id) REFERENCES shift (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE absensi (
  id BIGINT NOT NULL AUTO_INCREMENT,
  jadwal_shift_id BIGINT NOT NULL,
  jam_masuk DATETIME NOT NULL,
  jam_keluar DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_absensi_jadwal (jadwal_shift_id),
  KEY idx_absensi_masuk (jam_masuk),
  CONSTRAINT chk_absensi_jam CHECK (jam_keluar IS NULL OR jam_keluar > jam_masuk),
  CONSTRAINT fk_absensi_jadwal FOREIGN KEY (jadwal_shift_id) REFERENCES jadwal_shift (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE periode_gaji (
  id BIGINT NOT NULL AUTO_INCREMENT,
  tipe VARCHAR(10) NOT NULL,
  tanggal_mulai DATE NOT NULL,
  tanggal_selesai DATE NOT NULL,
  status VARCHAR(10) NOT NULL DEFAULT 'TERBUKA',
  ditutup_pada DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_periode_tipe_mulai (tipe, tanggal_mulai),
  KEY idx_periode_status (status),
  KEY idx_periode_rentang (tanggal_mulai, tanggal_selesai),
  CONSTRAINT chk_periode_tipe CHECK (tipe IN ('MINGGUAN', 'BULANAN')),
  CONSTRAINT chk_periode_status CHECK (status IN ('TERBUKA', 'DIHITUNG', 'DITUTUP')),
  CONSTRAINT chk_periode_tanggal CHECK (tanggal_selesai >= tanggal_mulai)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE kasbon (
  id BIGINT NOT NULL AUTO_INCREMENT,
  pekerja_id BIGINT NOT NULL,
  tanggal DATE NOT NULL,
  jumlah BIGINT NOT NULL,
  sisa BIGINT NOT NULL,
  keterangan VARCHAR(150) NULL,
  PRIMARY KEY (id),
  KEY idx_kasbon_pekerja (pekerja_id),
  KEY idx_kasbon_tanggal (tanggal),
  KEY idx_kasbon_sisa (pekerja_id, sisa),
  CONSTRAINT chk_kasbon_jumlah CHECK (jumlah > 0),
  CONSTRAINT chk_kasbon_sisa CHECK (sisa >= 0 AND sisa <= jumlah),
  CONSTRAINT fk_kasbon_pekerja FOREIGN KEY (pekerja_id) REFERENCES pekerja (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE slip_gaji (
  id BIGINT NOT NULL AUTO_INCREMENT,
  periode_id BIGINT NOT NULL,
  pekerja_id BIGINT NOT NULL,
  upah_pokok BIGINT NOT NULL,
  total_lembur BIGINT NOT NULL DEFAULT 0,
  potongan_telat BIGINT NOT NULL DEFAULT 0,
  total_koreksi BIGINT NOT NULL DEFAULT 0,
  gaji_bruto BIGINT NOT NULL,
  potongan_kasbon BIGINT NOT NULL DEFAULT 0,
  gaji_bersih BIGINT NOT NULL,
  hari_kerja_dibayar INT NOT NULL DEFAULT 0,
  hari_kerja_periode INT NULL,
  dibuat_pada DATETIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_slip_periode_pekerja (periode_id, pekerja_id),
  KEY idx_slip_periode (periode_id),
  KEY idx_slip_pekerja (pekerja_id),
  CONSTRAINT chk_slip_hari CHECK (hari_kerja_dibayar >= 0 AND (hari_kerja_periode IS NULL OR hari_kerja_periode >= 0)),
  CONSTRAINT chk_slip_potongan CHECK (potongan_telat >= 0 AND potongan_kasbon >= 0),
  CONSTRAINT fk_slip_periode FOREIGN KEY (periode_id) REFERENCES periode_gaji (id) ON DELETE RESTRICT,
  CONSTRAINT fk_slip_pekerja FOREIGN KEY (pekerja_id) REFERENCES pekerja (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE slip_gaji_detail (
  id BIGINT NOT NULL AUTO_INCREMENT,
  slip_id BIGINT NOT NULL,
  absensi_id BIGINT NOT NULL,
  tanggal DATE NOT NULL,
  upah_pokok_hari BIGINT NOT NULL,
  upah_per_jam BIGINT NOT NULL,
  menit_telat INT NOT NULL DEFAULT 0,
  potongan_telat BIGINT NOT NULL DEFAULT 0,
  menit_lembur INT NOT NULL DEFAULT 0,
  upah_lembur BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uq_detail_slip_absensi (slip_id, absensi_id),
  KEY idx_detail_slip (slip_id),
  KEY idx_detail_absensi (absensi_id),
  CONSTRAINT chk_detail_menit CHECK (menit_telat >= 0 AND menit_lembur >= 0),
  CONSTRAINT chk_detail_rupiah CHECK (upah_pokok_hari >= 0 AND upah_per_jam >= 0 AND potongan_telat >= 0 AND upah_lembur >= 0),
  CONSTRAINT fk_detail_slip FOREIGN KEY (slip_id) REFERENCES slip_gaji (id) ON DELETE RESTRICT,
  CONSTRAINT fk_detail_absensi FOREIGN KEY (absensi_id) REFERENCES absensi (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE potongan_kasbon (
  id BIGINT NOT NULL AUTO_INCREMENT,
  slip_id BIGINT NOT NULL,
  kasbon_id BIGINT NOT NULL,
  jumlah BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_potongan_slip_kasbon (slip_id, kasbon_id),
  KEY idx_potongan_slip (slip_id),
  KEY idx_potongan_kasbon (kasbon_id),
  CONSTRAINT chk_potongan_jumlah CHECK (jumlah > 0),
  CONSTRAINT fk_potongan_slip FOREIGN KEY (slip_id) REFERENCES slip_gaji (id) ON DELETE RESTRICT,
  CONSTRAINT fk_potongan_kasbon FOREIGN KEY (kasbon_id) REFERENCES kasbon (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE koreksi_gaji (
  id BIGINT NOT NULL AUTO_INCREMENT,
  pekerja_id BIGINT NOT NULL,
  periode_asal_id BIGINT NOT NULL,
  periode_terapan_id BIGINT NULL,
  jumlah BIGINT NOT NULL,
  keterangan VARCHAR(200) NOT NULL,
  dibuat_pada DATETIME NOT NULL,
  PRIMARY KEY (id),
  KEY idx_koreksi_pekerja (pekerja_id),
  KEY idx_koreksi_asal (periode_asal_id),
  KEY idx_koreksi_terapan (periode_terapan_id),
  CONSTRAINT chk_koreksi_jumlah CHECK (jumlah <> 0),
  CONSTRAINT fk_koreksi_pekerja FOREIGN KEY (pekerja_id) REFERENCES pekerja (id) ON DELETE RESTRICT,
  CONSTRAINT fk_koreksi_asal FOREIGN KEY (periode_asal_id) REFERENCES periode_gaji (id) ON DELETE RESTRICT,
  CONSTRAINT fk_koreksi_terapan FOREIGN KEY (periode_terapan_id) REFERENCES periode_gaji (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE pengguna (
  id BIGINT NOT NULL AUTO_INCREMENT,
  username VARCHAR(50) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(10) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_pengguna_username (username),
  CONSTRAINT chk_pengguna_role CHECK (role IN ('ADMIN', 'PEMILIK'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
