-- 000007: Data jam (informasional) pada slip gaji — model honorarium ala form Excel

ALTER TABLE payslips ADD COLUMN jam_mengajar INT;
ALTER TABLE payslips ADD COLUMN jam_tidak_hadir INT;
ALTER TABLE payslips ADD COLUMN hari_hadir INT;
