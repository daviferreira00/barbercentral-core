ALTER TABLE appointment DROP FOREIGN KEY fk_appointment_customer;

DROP TABLE IF EXISTS appointment_status_log;
DROP TABLE IF EXISTS appointment_payment;
DROP TABLE IF EXISTS customer;
