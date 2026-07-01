-- ============================================
-- SEED COMPLETO — BarberCentral
-- Execução direta, NÃO é migration.
-- Limpa dados existentes e repopula tudo.
-- ============================================

SET FOREIGN_KEY_CHECKS = 0;

-- Limpar tudo na ordem inversa de dependência
TRUNCATE TABLE loyalty_transaction;
TRUNCATE TABLE loyalty_card;
TRUNCATE TABLE loyalty_program;
TRUNCATE TABLE client_block_log;
TRUNCATE TABLE stock_movement;
TRUNCATE TABLE service_product;
TRUNCATE TABLE product;
TRUNCATE TABLE cash_transaction;
TRUNCATE TABLE cash_register;
TRUNCATE TABLE appointment_status_log;
TRUNCATE TABLE appointment_payment;
TRUNCATE TABLE appointment_service;
TRUNCATE TABLE appointment;
TRUNCATE TABLE blocked_slot;
TRUNCATE TABLE professional_schedule;
TRUNCATE TABLE professional_service;
TRUNCATE TABLE service;
TRUNCATE TABLE service_category;
TRUNCATE TABLE professional;
TRUNCATE TABLE client_config;
TRUNCATE TABLE auth_token;
TRUNCATE TABLE customer;
TRUNCATE TABLE client_user;
TRUNCATE TABLE client;
TRUNCATE TABLE platform_admin;
TRUNCATE TABLE plan;

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 1. PLANOS
-- ============================================
INSERT INTO plan (id, name, max_professionals, max_customers, billing_type, price, features_json, max_users, has_loyalty, has_stock, has_reports, has_online_booking, is_public) VALUES
('plan-basico',       'Básico',        2, 200,    'monthly',  0.00,   '{"agenda":true,"whatsapp":false}', 2, 0, 0, 0, 1, 1),
('plan-profissional', 'Profissional',  5, 500,    'monthly',  99.00,  '{"agenda":true,"whatsapp":true,"financeiro":true}', 5, 1, 1, 1, 1, 1),
('plan-premium',      'Premium',       9999, 999999, 'monthly', 199.00, '{"agenda":true,"whatsapp":true,"financeiro":true,"fidelidade":true}', 9999, 1, 1, 1, 1, 1);

-- ============================================
-- 2. PLATFORM ADMIN (senha: admin123)
-- ============================================
INSERT INTO platform_admin (id, name, email, password_hash) VALUES
('adm-01', 'Admin BarberCentral', 'admin@barbercentral.com.br', '$2a$10$16rxE.1lAn2zCCSHX3P6LeVNznY4lqWv6AgW28KyUKtEAjyOymRYK');

-- ============================================
-- 3. BARBEARIAS (CLIENTS)
-- ============================================
INSERT INTO client (id, plan_id, name, slug, status) VALUES
('cli-barber-modelo', 'plan-profissional', 'Barbearia Modelo',     'barbearia-modelo',  'active'),
('cli-corte-fino',    'plan-premium',      'Corte Fino Barbearia', 'corte-fino',        'active'),
('cli-teste-basico',  'plan-basico',       'Barba Rápida Express', 'barba-rapida',      'active');

-- ============================================
-- 4. USUÁRIOS DAS BARBEARIAS
-- ============================================
-- Todos com senha: senha123 → $2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K

-- Barbearia Modelo
INSERT INTO client_user (id, client_id, name, email, password_hash, role, status) VALUES
('usr-joao',    'cli-barber-modelo', 'João Barbeiro',    'joao@barberiamodelo.com',    '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'owner',          'active'),
('usr-marcos',  'cli-barber-modelo', 'Marcos Cabeleireiro', 'marcos@barberiamodelo.com', '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'professional',   'active'),
('usr-tiago',   'cli-barber-modelo', 'Tiago Barbeiro',   'tiago@barberiamodelo.com',   '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'professional',   'active'),
('usr-rodolfo', 'cli-barber-modelo', 'Rodolfo Estilista','rodolfo@barberiamodelo.com','$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'professional',   'active'),
('usr-recep',   'cli-barber-modelo', 'Maria Recepção',   'maria@barberiamodelo.com',   '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'receptionist',   'active');

-- Corte Fino
INSERT INTO client_user (id, client_id, name, email, password_hash, role, status) VALUES
('usr-carlos',  'cli-corte-fino', 'Carlos Lopes',  'carlos@cortefino.com',  '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'owner',        'active'),
('usr-roberto', 'cli-corte-fino', 'Roberto Silva', 'roberto@cortefino.com', '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'professional', 'active');

-- Barba Rápida
INSERT INTO client_user (id, client_id, name, email, password_hash, role, status) VALUES
('usr-andre', 'cli-teste-basico', 'André Bastos', 'andre@barbarapida.com', '$2a$10$NI0.3ybMrRJNdvIpuUVJsegMFhiG/JIByXNVRZYux2L9LN6PfQG9K', 'owner', 'active');

-- ============================================
-- 5. CLIENT CONFIG (Identidade Visual e Configurações)
-- ============================================
INSERT INTO client_config (client_id, logo_url, color_primary, color_secondary, font_family, address, neighborhood, city, state, phone, whatsapp, instagram, timezone, cancellation_policy_hours, booking_requires_login, min_advance_hours, max_advance_days, interval_between_minutes, active) VALUES
('cli-barber-modelo', NULL, '#0f172a', '#d97706', 'Inter', 'Rua das Palmeiras, 150', 'Centro', 'São Paulo', 'SP', '11999999999', '11999999999', 'barbeariamodelo', 'America/Sao_Paulo', 2, 0, 1, 30, 15, 1),
('cli-corte-fino',    NULL, '#1e293b', '#ef4444', 'Poppins', 'Av. Brasil, 2000', 'Jardins', 'São Paulo', 'SP', '11988887777', '11988887777', 'cortefino', 'America/Sao_Paulo', 4, 0, 2, 14, 10, 1),
('cli-teste-basico',  NULL, '#111827', '#10b981', 'Roboto', 'Rua XV de Novembro, 45', 'Centro', 'Curitiba', 'PR', '41977776666', '41977776666', 'barbarapida', 'America/Sao_Paulo', 1, 0, 1, 7, 0, 1);

-- ============================================
-- 6. PROFISSIONAIS
-- ============================================
-- Barbearia Modelo (3 profissionais)
INSERT INTO professional (id, client_id, user_id, name, bio, photo_url, status) VALUES
('prof-marcos',  'cli-barber-modelo', 'usr-marcos',  'Marcos Cabeleireiro', 'Especialista em cortes clássicos e tesoura com 10 anos de experiência.', NULL, 'active'),
('prof-tiago',   'cli-barber-modelo', 'usr-tiago',   'Tiago Barbeiro',     'Mestre na navalha e especialista em barboterapia e degradês modernos.',    NULL, 'active'),
('prof-rodolfo', 'cli-barber-modelo', 'usr-rodolfo', 'Rodolfo Estilista',  'Cabeleireiro com foco em design de estilos, químicas e hidratações.',     NULL, 'active');

-- Corte Fino (1 profissional)
INSERT INTO professional (id, client_id, user_id, name, bio, photo_url, status) VALUES
('prof-roberto', 'cli-corte-fino', 'usr-roberto', 'Roberto Silva', 'Barbeiro com 15 anos de carreira. Especialista em cortes americanos.', NULL, 'active');

-- Barba Rápida (dono é o profissional)
INSERT INTO professional (id, client_id, user_id, name, bio, photo_url, status) VALUES
('prof-andre', 'cli-teste-basico', 'usr-andre', 'André Bastos', 'Barbeiro único do salão. Atende com hora marcada e walk-in.', NULL, 'active');

-- ============================================
-- 7. CATEGORIAS DE SERVIÇOS
-- ============================================
INSERT INTO service_category (id, client_id, name) VALUES
('cat-cabelo-1',  'cli-barber-modelo', 'Cabelo'),
('cat-barba-1',   'cli-barber-modelo', 'Barba'),
('cat-tratam-1',  'cli-barber-modelo', 'Tratamentos'),
('cat-cabelo-2',  'cli-corte-fino',    'Cabelo'),
('cat-barba-2',   'cli-corte-fino',    'Barba'),
('cat-cabelo-3',  'cli-teste-basico',  'Cortes'),
('cat-barba-3',   'cli-teste-basico',  'Barbas');

-- ============================================
-- 8. SERVIÇOS
-- ============================================
-- Barbearia Modelo
INSERT INTO service (id, client_id, category_id, name, description, duration_minutes, price, active) VALUES
('srv-degrade',     'cli-barber-modelo', 'cat-cabelo-1', 'Corte Degradê',          'Corte moderno com acabamento em degradê',               30, 45.00, 1),
('srv-barba',       'cli-barber-modelo', 'cat-barba-1',  'Barba Simples',          'Aparar barba com toalha quente e navalha',              20, 30.00, 1),
('srv-combo',       'cli-barber-modelo', 'cat-cabelo-1', 'Corte + Barba',          'Combo completo corte de cabelo e barba premium',        50, 65.00, 1),
('srv-hidratacao',  'cli-barber-modelo', 'cat-tratam-1', 'Hidratação Capilar',     'Tratamento de hidratação e brilho dos fios',            40, 40.00, 1),
('srv-pigmenta',    'cli-barber-modelo', 'cat-barba-1',  'Pigmentação de Barba',   'Camuflagem de fios brancos e preenchimento de falhas',  30, 35.00, 1),
('srv-tesoura',     'cli-barber-modelo', 'cat-cabelo-1', 'Corte Tesoura',          'Corte clássico com tesoura e acabamento refinado',      40, 55.00, 1),
('srv-sobrancelha', 'cli-barber-modelo', 'cat-tratam-1', 'Design de Sobrancelha',  'Alinhamento e design masculino com pinça e navalha',    15, 20.00, 1),
('srv-relaxamento', 'cli-barber-modelo', 'cat-tratam-1', 'Relaxamento Capilar',    'Alisamento e relaxamento para cabelos crespos/cacheados', 60, 80.00, 1);

-- Corte Fino
INSERT INTO service (id, client_id, category_id, name, description, duration_minutes, price, active) VALUES
('srv-cf-corte',   'cli-corte-fino', 'cat-cabelo-2', 'Corte Americano',   'Corte clássico americano com fade',           35, 50.00, 1),
('srv-cf-barba',   'cli-corte-fino', 'cat-barba-2',  'Barba Completa',    'Barba feita com navalha e toalha quente',     25, 35.00, 1),
('srv-cf-combo',   'cli-corte-fino', 'cat-cabelo-2', 'Corte + Barba',     'Combo com acabamento premium',                55, 75.00, 1);

-- Barba Rápida
INSERT INTO service (id, client_id, category_id, name, description, duration_minutes, price, active) VALUES
('srv-br-corte',  'cli-teste-basico', 'cat-cabelo-3', 'Corte Express',   'Corte rápido de máquina e tesoura',  20, 30.00, 1),
('srv-br-barba',  'cli-teste-basico', 'cat-barba-3',  'Barba Express',   'Barba na navalha expressa',          15, 20.00, 1);

-- ============================================
-- 9. VÍNCULO PROFISSIONAL ↔ SERVIÇO
-- ============================================
-- Barbearia Modelo
INSERT INTO professional_service (professional_id, service_id, client_id) VALUES
('prof-marcos', 'srv-degrade',     'cli-barber-modelo'),
('prof-marcos', 'srv-combo',       'cli-barber-modelo'),
('prof-marcos', 'srv-hidratacao',  'cli-barber-modelo'),
('prof-marcos', 'srv-tesoura',     'cli-barber-modelo'),
('prof-marcos', 'srv-relaxamento', 'cli-barber-modelo'),
('prof-tiago',  'srv-degrade',     'cli-barber-modelo'),
('prof-tiago',  'srv-barba',       'cli-barber-modelo'),
('prof-tiago',  'srv-combo',       'cli-barber-modelo'),
('prof-tiago',  'srv-pigmenta',    'cli-barber-modelo'),
('prof-tiago',  'srv-sobrancelha', 'cli-barber-modelo'),
('prof-rodolfo','srv-degrade',     'cli-barber-modelo'),
('prof-rodolfo','srv-hidratacao',  'cli-barber-modelo'),
('prof-rodolfo','srv-tesoura',     'cli-barber-modelo'),
('prof-rodolfo','srv-relaxamento', 'cli-barber-modelo');

-- Corte Fino
INSERT INTO professional_service (professional_id, service_id, client_id) VALUES
('prof-roberto', 'srv-cf-corte', 'cli-corte-fino'),
('prof-roberto', 'srv-cf-barba', 'cli-corte-fino'),
('prof-roberto', 'srv-cf-combo', 'cli-corte-fino');

-- Barba Rápida
INSERT INTO professional_service (professional_id, service_id, client_id) VALUES
('prof-andre', 'srv-br-corte', 'cli-teste-basico'),
('prof-andre', 'srv-br-barba', 'cli-teste-basico');

-- ============================================
-- 10. GRADES DE HORÁRIOS (Seg-Sáb 09:00-19:00)
-- ============================================
INSERT INTO professional_schedule (id, professional_id, client_id, weekday, start_time, end_time, enabled) VALUES
-- Marcos (Seg-Sáb)
('sch-marcos-1','prof-marcos','cli-barber-modelo',1,'09:00','19:00',1),
('sch-marcos-2','prof-marcos','cli-barber-modelo',2,'09:00','19:00',1),
('sch-marcos-3','prof-marcos','cli-barber-modelo',3,'09:00','19:00',1),
('sch-marcos-4','prof-marcos','cli-barber-modelo',4,'09:00','19:00',1),
('sch-marcos-5','prof-marcos','cli-barber-modelo',5,'09:00','19:00',1),
('sch-marcos-6','prof-marcos','cli-barber-modelo',6,'09:00','17:00',1),
-- Tiago (Seg-Sáb)
('sch-tiago-1','prof-tiago','cli-barber-modelo',1,'10:00','20:00',1),
('sch-tiago-2','prof-tiago','cli-barber-modelo',2,'10:00','20:00',1),
('sch-tiago-3','prof-tiago','cli-barber-modelo',3,'10:00','20:00',1),
('sch-tiago-4','prof-tiago','cli-barber-modelo',4,'10:00','20:00',1),
('sch-tiago-5','prof-tiago','cli-barber-modelo',5,'10:00','20:00',1),
('sch-tiago-6','prof-tiago','cli-barber-modelo',6,'10:00','17:00',1),
-- Rodolfo (Ter-Sáb)
('sch-rodolfo-2','prof-rodolfo','cli-barber-modelo',2,'08:00','18:00',1),
('sch-rodolfo-3','prof-rodolfo','cli-barber-modelo',3,'08:00','18:00',1),
('sch-rodolfo-4','prof-rodolfo','cli-barber-modelo',4,'08:00','18:00',1),
('sch-rodolfo-5','prof-rodolfo','cli-barber-modelo',5,'08:00','18:00',1),
('sch-rodolfo-6','prof-rodolfo','cli-barber-modelo',6,'08:00','15:00',1),
-- Roberto (Corte Fino - Seg-Sex)
('sch-roberto-1','prof-roberto','cli-corte-fino',1,'09:00','18:00',1),
('sch-roberto-2','prof-roberto','cli-corte-fino',2,'09:00','18:00',1),
('sch-roberto-3','prof-roberto','cli-corte-fino',3,'09:00','18:00',1),
('sch-roberto-4','prof-roberto','cli-corte-fino',4,'09:00','18:00',1),
('sch-roberto-5','prof-roberto','cli-corte-fino',5,'09:00','18:00',1),
-- André (Barba Rápida - Seg-Sáb)
('sch-andre-1','prof-andre','cli-teste-basico',1,'08:00','20:00',1),
('sch-andre-2','prof-andre','cli-teste-basico',2,'08:00','20:00',1),
('sch-andre-3','prof-andre','cli-teste-basico',3,'08:00','20:00',1),
('sch-andre-4','prof-andre','cli-teste-basico',4,'08:00','20:00',1),
('sch-andre-5','prof-andre','cli-teste-basico',5,'08:00','20:00',1),
('sch-andre-6','prof-andre','cli-teste-basico',6,'08:00','16:00',1);

-- ============================================
-- 11. CLIENTES (CRM — CUSTOMERS)
-- ============================================
-- Barbearia Modelo (15 clientes)
INSERT INTO customer (id, client_id, name, phone, email, cpf, birth_date, notes) VALUES
('cust-pedro',     'cli-barber-modelo', 'Pedro Alvares',     '11911111111', 'pedro@email.com',     '123.456.789-01', '1990-05-15', 'Cliente VIP, prefere café expresso.'),
('cust-lucas',     'cli-barber-modelo', 'Lucas Silva',       '11922222222', 'lucas@email.com',     '234.567.890-12', '1985-11-20', 'Cabelo afro, costuma fazer degradê na zero.'),
('cust-mariana',   'cli-barber-modelo', 'Mariana Costa',     '11933333333', 'mariana@email.com',   NULL,              '1998-02-28', 'Faz alinhamento capilar bimestral.'),
('cust-carlos',    'cli-barber-modelo', 'Carlos Eduardo',    '11944444444', 'carlos.ed@email.com', '345.678.901-23', '1979-07-04', 'Sempre pontual.'),
('cust-gabriel',   'cli-barber-modelo', 'Gabriel Souza',     '11955555555', 'gabriel@email.com',   NULL,              '1993-10-12', 'Vem a cada 15 dias.'),
('cust-bruno',     'cli-barber-modelo', 'Bruno Henrique',    '11966666666', 'bruno@email.com',     '456.789.012-34', '1991-03-08', NULL),
('cust-felipe',    'cli-barber-modelo', 'Felipe Santos',     '11977777777', 'felipe@email.com',    NULL,              '1995-12-25', 'Prefere cortes clássicos na tesoura.'),
('cust-thiago',    'cli-barber-modelo', 'Thiago Lima',       '11988888888', 'thiago@email.com',    '567.890.123-45', '1988-08-30', NULL),
('cust-rodrigo',   'cli-barber-modelo', 'Rodrigo Alves',     '11999999990', 'rodrigo@email.com',   NULL,              '1994-06-18', NULL),
('cust-matheus',   'cli-barber-modelo', 'Matheus Oliveira',  '11910101010', 'matheus@email.com',   '678.901.234-56', '1987-04-02', 'Corte e barba terapia.'),
('cust-daniel',    'cli-barber-modelo', 'Daniel Pires',      '11911011101', 'daniel@email.com',    NULL,              '1992-09-09', NULL),
('cust-rafael',    'cli-barber-modelo', 'Rafael Dutra',      '11912011201', 'rafael@email.com',    '789.012.345-67', '2000-01-01', NULL),
('cust-alexandre', 'cli-barber-modelo', 'Alexandre Borges',  '11913011301', 'alexandre@email.com', NULL,              '1983-05-30', NULL),
('cust-leonardo',  'cli-barber-modelo', 'Leonardo Cruz',     '11914011401', 'leonardo@email.com',  '890.123.456-78', '1996-08-12', NULL),
('cust-gustavo',   'cli-barber-modelo', 'Gustavo Paiva',     '11915011501', 'gustavo@email.com',   NULL,              '1997-12-03', NULL);

-- Corte Fino (5 clientes)
INSERT INTO customer (id, client_id, name, phone, email, cpf, birth_date, notes) VALUES
('cust-cf-1', 'cli-corte-fino', 'Ricardo Menezes',   '11981111111', 'ricardo.m@email.com', NULL, '1990-03-10', 'Prefere horário da manhã.'),
('cust-cf-2', 'cli-corte-fino', 'Fábio Costa',       '11982222222', 'fabio.c@email.com',   NULL, '1988-07-22', NULL),
('cust-cf-3', 'cli-corte-fino', 'Henrique Tavares',  '11983333333', 'henrique.t@email.com',NULL, '1995-01-15', NULL),
('cust-cf-4', 'cli-corte-fino', 'Vinícius Ramos',    '11984444444', 'vinicius.r@email.com',NULL, '1992-11-05', 'Cliente regular.'),
('cust-cf-5', 'cli-corte-fino', 'Diego Fernandes',   '11985555555', 'diego.f@email.com',   NULL, '1999-06-30', NULL);

-- Barba Rápida (3 clientes)
INSERT INTO customer (id, client_id, name, phone, email, cpf, birth_date, notes) VALUES
('cust-br-1', 'cli-teste-basico', 'Wagner Nunes',     '41971111111', 'wagner@email.com', NULL, '1991-04-20', NULL),
('cust-br-2', 'cli-teste-basico', 'Eduardo Campos',   '41972222222', 'eduardo@email.com',NULL, '1986-08-14', NULL),
('cust-br-3', 'cli-teste-basico', 'Sérgio Machado',   '41973333333', 'sergio@email.com', NULL, '1993-12-01', NULL);

-- ============================================
-- 12. BLOQUEIOS DE AGENDA (EXEMPLO)
-- ============================================
INSERT INTO blocked_slot (id, client_id, professional_id, date, start_time, end_time, reason, created_by) VALUES
('blk-1', 'cli-barber-modelo', 'prof-marcos', CURDATE(), '12:00', '13:00', 'Horário de Almoço', 'usr-joao'),
('blk-2', 'cli-barber-modelo', 'prof-tiago',  CURDATE(), '12:00', '13:30', 'Almoço', 'usr-joao'),
('blk-3', 'cli-barber-modelo', NULL,          DATE_ADD(CURDATE(), INTERVAL 3 DAY), '08:00', '10:00', 'Dedetização do salão', 'usr-joao');

-- ============================================
-- 13. AGENDAMENTOS (semana atual e passada)
-- ============================================

-- == AGENDAMENTOS PASSADOS (CONCLUÍDOS — semana anterior) ==
INSERT INTO appointment (id, client_id, professional_id, customer_id, date, start_time, end_time, status, notes, source, cancel_token, customer_name, customer_phone, customer_email, reminder_sent) VALUES
-- 5 dias atrás
('apt-h1', 'cli-barber-modelo', 'prof-marcos', 'cust-pedro',   DATE_SUB(CURDATE(), INTERVAL 5 DAY), '09:30', '10:00', 'completed', 'Corte degradê',           'panel',  NULL, 'Pedro Alvares',    '11911111111', 'pedro@email.com',    0),
('apt-h2', 'cli-barber-modelo', 'prof-tiago',  'cust-lucas',   DATE_SUB(CURDATE(), INTERVAL 5 DAY), '10:30', '10:50', 'completed', 'Barba simples',           'panel',  NULL, 'Lucas Silva',      '11922222222', 'lucas@email.com',    0),
('apt-h3', 'cli-barber-modelo', 'prof-marcos', 'cust-carlos',  DATE_SUB(CURDATE(), INTERVAL 5 DAY), '14:00', '14:50', 'completed', 'Combo corte e barba',     'panel',  NULL, 'Carlos Eduardo',   '11944444444', 'carlos.ed@email.com',0),
('apt-h4', 'cli-barber-modelo', 'prof-rodolfo','cust-gabriel',  DATE_SUB(CURDATE(), INTERVAL 5 DAY), '15:00', '15:40', 'completed', 'Hidratação',              'panel',  NULL, 'Gabriel Souza',    '11955555555', 'gabriel@email.com',  0),
-- 4 dias atrás
('apt-h5', 'cli-barber-modelo', 'prof-tiago',  'cust-felipe',  DATE_SUB(CURDATE(), INTERVAL 4 DAY), '10:00', '10:50', 'completed', 'Corte e barba',           'panel',  NULL, 'Felipe Santos',    '11977777777', 'felipe@email.com',   0),
('apt-h6', 'cli-barber-modelo', 'prof-marcos', 'cust-bruno',   DATE_SUB(CURDATE(), INTERVAL 4 DAY), '11:00', '11:30', 'completed', 'Corte degradê',           'online', NULL, 'Bruno Henrique',   '11966666666', 'bruno@email.com',    0),
('apt-h7', 'cli-barber-modelo', 'prof-rodolfo','cust-thiago',  DATE_SUB(CURDATE(), INTERVAL 4 DAY), '14:00', '15:00', 'completed', 'Relaxamento capilar',     'panel',  NULL, 'Thiago Lima',      '11988888888', 'thiago@email.com',   0),
-- 3 dias atrás
('apt-h8', 'cli-barber-modelo', 'prof-marcos', 'cust-pedro',   DATE_SUB(CURDATE(), INTERVAL 3 DAY), '09:00', '09:30', 'completed', 'Retoque de degradê',      'panel',  NULL, 'Pedro Alvares',    '11911111111', 'pedro@email.com',    0),
('apt-h9', 'cli-barber-modelo', 'prof-tiago',  'cust-matheus', DATE_SUB(CURDATE(), INTERVAL 3 DAY), '11:00', '11:50', 'completed', 'Corte + barba combo',     'panel',  NULL, 'Matheus Oliveira', '11910101010', 'matheus@email.com',  0),
('apt-h10','cli-barber-modelo', 'prof-tiago',  'cust-rodrigo', DATE_SUB(CURDATE(), INTERVAL 3 DAY), '14:00', '14:30', 'cancelled', 'Cancelou de última hora',  'online', NULL, 'Rodrigo Alves',    '11999999990', 'rodrigo@email.com',  0),
-- 2 dias atrás
('apt-h11','cli-barber-modelo', 'prof-marcos', 'cust-daniel',   DATE_SUB(CURDATE(), INTERVAL 2 DAY), '10:00', '10:30', 'completed', 'Corte tesoura',            'panel',  NULL, 'Daniel Pires',     '11911011101', 'daniel@email.com',   0),
('apt-h12','cli-barber-modelo', 'prof-rodolfo','cust-rafael',   DATE_SUB(CURDATE(), INTERVAL 2 DAY), '14:00', '14:40', 'no_show',   'Não compareceu',           'online', NULL, 'Rafael Dutra',     '11912011201', 'rafael@email.com',   0),
('apt-h13','cli-barber-modelo', 'prof-tiago',  'cust-leonardo', DATE_SUB(CURDATE(), INTERVAL 2 DAY), '15:00', '15:30', 'completed', 'Pigmentação de barba',     'panel',  NULL, 'Leonardo Cruz',    '11914011401', 'leonardo@email.com', 0),
-- 1 dia atrás (ontem)
('apt-h14','cli-barber-modelo', 'prof-marcos', 'cust-gustavo',  DATE_SUB(CURDATE(), INTERVAL 1 DAY), '09:00', '09:50', 'completed', 'Combo corte + barba',      'panel',  NULL, 'Gustavo Paiva',    '11915011501', 'gustavo@email.com',  0),
('apt-h15','cli-barber-modelo', 'prof-tiago',  'cust-pedro',    DATE_SUB(CURDATE(), INTERVAL 1 DAY), '10:00', '10:30', 'completed', 'Corte degradê rápido',     'panel',  NULL, 'Pedro Alvares',    '11911111111', 'pedro@email.com',    0),
('apt-h16','cli-barber-modelo', 'prof-rodolfo','cust-alexandre',DATE_SUB(CURDATE(), INTERVAL 1 DAY), '14:00', '14:40', 'completed', 'Hidratação pós-química',   'panel',  NULL, 'Alexandre Borges', '11913011301', 'alexandre@email.com',0);

-- == AGENDAMENTOS DE HOJE ==
INSERT INTO appointment (id, client_id, professional_id, customer_id, date, start_time, end_time, status, notes, source, cancel_token, customer_name, customer_phone, customer_email, reminder_sent) VALUES
('apt-t1', 'cli-barber-modelo', 'prof-marcos', 'cust-lucas',   CURDATE(), '09:00', '09:30', 'completed',   'Corte degradê matinal',   'panel',  NULL, 'Lucas Silva',     '11922222222', 'lucas@email.com',    0),
('apt-t2', 'cli-barber-modelo', 'prof-tiago',  'cust-mariana', CURDATE(), '10:00', '10:40', 'in_progress', 'Hidratação capilar',      'online', NULL, 'Mariana Costa',   '11933333333', 'mariana@email.com',  1),
('apt-t3', 'cli-barber-modelo', 'prof-marcos', 'cust-bruno',   CURDATE(), '11:00', '11:50', 'confirmed',   'Corte + barba completo',  'panel',  NULL, 'Bruno Henrique',  '11966666666', 'bruno@email.com',    0),
('apt-t4', 'cli-barber-modelo', 'prof-rodolfo','cust-carlos',  CURDATE(), '14:00', '14:40', 'confirmed',   'Corte tesoura',           'panel',  NULL, 'Carlos Eduardo',  '11944444444', 'carlos.ed@email.com',0),
('apt-t5', 'cli-barber-modelo', 'prof-tiago',  'cust-gabriel', CURDATE(), '15:00', '15:30', 'pending',     'Pigmentação de barba',    'online', NULL, 'Gabriel Souza',   '11955555555', 'gabriel@email.com',  0),
('apt-t6', 'cli-barber-modelo', 'prof-marcos', 'cust-thiago',  CURDATE(), '16:00', '17:00', 'pending',     'Relaxamento capilar',     'panel',  NULL, 'Thiago Lima',     '11988888888', 'thiago@email.com',   0);

-- == AGENDAMENTOS FUTUROS (AMANHÃ E DEPOIS) ==
INSERT INTO appointment (id, client_id, professional_id, customer_id, date, start_time, end_time, status, notes, source, cancel_token, customer_name, customer_phone, customer_email, reminder_sent) VALUES
('apt-f1', 'cli-barber-modelo', 'prof-marcos', 'cust-pedro',    DATE_ADD(CURDATE(), INTERVAL 1 DAY), '09:30', '10:00', 'confirmed', 'Retoque degradê',         'online', NULL, 'Pedro Alvares',   '11911111111', 'pedro@email.com',   0),
('apt-f2', 'cli-barber-modelo', 'prof-tiago',  'cust-felipe',   DATE_ADD(CURDATE(), INTERVAL 1 DAY), '10:00', '10:50', 'confirmed', 'Corte + barba',           'panel',  NULL, 'Felipe Santos',   '11977777777', 'felipe@email.com',  0),
('apt-f3', 'cli-barber-modelo', 'prof-rodolfo','cust-matheus',  DATE_ADD(CURDATE(), INTERVAL 1 DAY), '14:00', '15:00', 'pending',   'Relaxamento',             'panel',  NULL, 'Matheus Oliveira','11910101010', 'matheus@email.com', 0),
('apt-f4', 'cli-barber-modelo', 'prof-marcos', 'cust-daniel',   DATE_ADD(CURDATE(), INTERVAL 2 DAY), '11:00', '11:30', 'confirmed', 'Corte degradê',           'online', NULL, 'Daniel Pires',    '11911011101', 'daniel@email.com',  0),
('apt-f5', 'cli-barber-modelo', 'prof-tiago',  'cust-gustavo',  DATE_ADD(CURDATE(), INTERVAL 2 DAY), '15:00', '15:50', 'pending',   'Combo completo',          'panel',  NULL, 'Gustavo Paiva',   '11915011501', 'gustavo@email.com', 0);

-- Corte Fino
INSERT INTO appointment (id, client_id, professional_id, customer_id, date, start_time, end_time, status, notes, source, cancel_token, customer_name, customer_phone, customer_email, reminder_sent) VALUES
('apt-cf-1', 'cli-corte-fino', 'prof-roberto', 'cust-cf-1', DATE_SUB(CURDATE(), INTERVAL 2 DAY), '10:00', '10:35', 'completed', 'Corte americano',    'panel', NULL, 'Ricardo Menezes', '11981111111', 'ricardo.m@email.com', 0),
('apt-cf-2', 'cli-corte-fino', 'prof-roberto', 'cust-cf-2', DATE_SUB(CURDATE(), INTERVAL 1 DAY), '11:00', '11:55', 'completed', 'Corte + barba',      'panel', NULL, 'Fábio Costa',     '11982222222', 'fabio.c@email.com',   0),
('apt-cf-3', 'cli-corte-fino', 'prof-roberto', 'cust-cf-3', CURDATE(), '09:30', '10:05', 'confirmed',     'Corte americano',    'online',NULL, 'Henrique Tavares','11983333333', 'henrique.t@email.com',0);

-- ============================================
-- 14. SERVIÇOS DOS AGENDAMENTOS
-- ============================================
INSERT INTO appointment_service (id, appointment_id, service_id, price, duration_minutes) VALUES
-- Passados (Barbearia Modelo)
('as-h1',  'apt-h1',  'srv-degrade',    45.00, 30),
('as-h2',  'apt-h2',  'srv-barba',      30.00, 20),
('as-h3',  'apt-h3',  'srv-combo',      65.00, 50),
('as-h4',  'apt-h4',  'srv-hidratacao', 40.00, 40),
('as-h5',  'apt-h5',  'srv-combo',      65.00, 50),
('as-h6',  'apt-h6',  'srv-degrade',    45.00, 30),
('as-h7',  'apt-h7',  'srv-relaxamento',80.00, 60),
('as-h8',  'apt-h8',  'srv-degrade',    45.00, 30),
('as-h9',  'apt-h9',  'srv-combo',      65.00, 50),
('as-h10', 'apt-h10', 'srv-degrade',    45.00, 30),
('as-h11', 'apt-h11', 'srv-tesoura',    55.00, 40),
('as-h12', 'apt-h12', 'srv-hidratacao', 40.00, 40),
('as-h13', 'apt-h13', 'srv-pigmenta',   35.00, 30),
('as-h14', 'apt-h14', 'srv-combo',      65.00, 50),
('as-h15', 'apt-h15', 'srv-degrade',    45.00, 30),
('as-h16', 'apt-h16', 'srv-hidratacao', 40.00, 40),
-- Hoje
('as-t1',  'apt-t1',  'srv-degrade',    45.00, 30),
('as-t2',  'apt-t2',  'srv-hidratacao', 40.00, 40),
('as-t3',  'apt-t3',  'srv-combo',      65.00, 50),
('as-t4',  'apt-t4',  'srv-tesoura',    55.00, 40),
('as-t5',  'apt-t5',  'srv-pigmenta',   35.00, 30),
('as-t6',  'apt-t6',  'srv-relaxamento',80.00, 60),
-- Futuros
('as-f1',  'apt-f1',  'srv-degrade',    45.00, 30),
('as-f2',  'apt-f2',  'srv-combo',      65.00, 50),
('as-f3',  'apt-f3',  'srv-relaxamento',80.00, 60),
('as-f4',  'apt-f4',  'srv-degrade',    45.00, 30),
('as-f5',  'apt-f5',  'srv-combo',      65.00, 50),
-- Corte Fino
('as-cf-1','apt-cf-1','srv-cf-corte',   50.00, 35),
('as-cf-2','apt-cf-2','srv-cf-combo',   75.00, 55),
('as-cf-3','apt-cf-3','srv-cf-corte',   50.00, 35);

-- ============================================
-- 15. LOGS DE STATUS DOS AGENDAMENTOS
-- ============================================
INSERT INTO appointment_status_log (id, appointment_id, from_status, to_status, changed_by, notes) VALUES
('log-1',  'apt-h1',  'pending',   'confirmed', 'usr-joao',    'Confirmado pelo painel'),
('log-2',  'apt-h1',  'confirmed', 'completed', 'usr-marcos',  'Serviço concluído'),
('log-3',  'apt-h2',  'pending',   'completed', 'usr-tiago',   'Atendimento express'),
('log-4',  'apt-h3',  'pending',   'confirmed', 'usr-joao',    'Confirmado'),
('log-5',  'apt-h3',  'confirmed', 'completed', 'usr-marcos',  'Finalizado'),
('log-6',  'apt-h5',  'pending',   'completed', 'usr-tiago',   'Atendimento direto'),
('log-7',  'apt-h6',  'pending',   'completed', 'usr-marcos',  'Concluído'),
('log-8',  'apt-h10', 'pending',   'cancelled', 'usr-tiago',   'Cliente cancelou por telefone'),
('log-9',  'apt-h12', 'pending',   'no_show',   'usr-rodolfo', 'Não compareceu, ligou depois'),
('log-10', 'apt-t1',  'pending',   'confirmed', 'usr-marcos',  'Confirmado'),
('log-11', 'apt-t1',  'confirmed', 'completed', 'usr-marcos',  'Concluído'),
('log-12', 'apt-t2',  'pending',   'confirmed', 'usr-tiago',   'Confirmação online'),
('log-13', 'apt-t2',  'confirmed', 'in_progress','usr-tiago',  'Iniciado atendimento');

-- ============================================
-- 16. PAGAMENTOS DOS AGENDAMENTOS CONCLUÍDOS
-- ============================================
INSERT INTO appointment_payment (id, appointment_id, client_id, amount, method, status, paid_at, notes) VALUES
('pay-h1',  'apt-h1',  'cli-barber-modelo', 45.00,  'pix',         'paid', DATE_SUB(NOW(), INTERVAL 5 DAY), NULL),
('pay-h2',  'apt-h2',  'cli-barber-modelo', 30.00,  'cash',        'paid', DATE_SUB(NOW(), INTERVAL 5 DAY), NULL),
('pay-h3',  'apt-h3',  'cli-barber-modelo', 65.00,  'card_credit', 'paid', DATE_SUB(NOW(), INTERVAL 5 DAY), NULL),
('pay-h4',  'apt-h4',  'cli-barber-modelo', 40.00,  'pix',         'paid', DATE_SUB(NOW(), INTERVAL 5 DAY), NULL),
('pay-h5',  'apt-h5',  'cli-barber-modelo', 65.00,  'card_debit',  'paid', DATE_SUB(NOW(), INTERVAL 4 DAY), NULL),
('pay-h6',  'apt-h6',  'cli-barber-modelo', 45.00,  'pix',         'paid', DATE_SUB(NOW(), INTERVAL 4 DAY), NULL),
('pay-h7',  'apt-h7',  'cli-barber-modelo', 80.00,  'card_credit', 'paid', DATE_SUB(NOW(), INTERVAL 4 DAY), NULL),
('pay-h8',  'apt-h8',  'cli-barber-modelo', 45.00,  'cash',        'paid', DATE_SUB(NOW(), INTERVAL 3 DAY), NULL),
('pay-h9',  'apt-h9',  'cli-barber-modelo', 65.00,  'pix',         'paid', DATE_SUB(NOW(), INTERVAL 3 DAY), NULL),
('pay-h11', 'apt-h11', 'cli-barber-modelo', 55.00,  'card_debit',  'paid', DATE_SUB(NOW(), INTERVAL 2 DAY), NULL),
('pay-h13', 'apt-h13', 'cli-barber-modelo', 35.00,  'cash',        'paid', DATE_SUB(NOW(), INTERVAL 2 DAY), NULL),
('pay-h14', 'apt-h14', 'cli-barber-modelo', 65.00,  'pix',         'paid', DATE_SUB(NOW(), INTERVAL 1 DAY), NULL),
('pay-h15', 'apt-h15', 'cli-barber-modelo', 45.00,  'card_credit', 'paid', DATE_SUB(NOW(), INTERVAL 1 DAY), NULL),
('pay-h16', 'apt-h16', 'cli-barber-modelo', 40.00,  'pix',         'paid', DATE_SUB(NOW(), INTERVAL 1 DAY), NULL),
('pay-t1',  'apt-t1',  'cli-barber-modelo', 45.00,  'cash',        'paid', NOW(),                            NULL),
-- Corte Fino
('pay-cf1', 'apt-cf-1','cli-corte-fino',    50.00,  'pix',         'paid', DATE_SUB(NOW(), INTERVAL 2 DAY), NULL),
('pay-cf2', 'apt-cf-2','cli-corte-fino',    75.00,  'card_credit', 'paid', DATE_SUB(NOW(), INTERVAL 1 DAY), NULL);

-- ============================================
-- 17. CAIXAS (CASH REGISTERS)
-- ============================================
-- Barbearia Modelo — caixa de ontem (fechado) e caixa de hoje (aberto)
INSERT INTO cash_register (id, client_id, opened_by, opened_at, closed_by, closed_at, opening_balance, closing_balance, notes, status) VALUES
('cr-ontem',  'cli-barber-modelo', 'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 8 HOUR, 'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 20 HOUR, 100.00, 520.00, 'Caixa fechado sem divergências', 'closed'),
('cr-hoje',   'cli-barber-modelo', 'usr-recep', CURDATE() + INTERVAL 8 HOUR, NULL, NULL, 150.00, NULL, NULL, 'open');

-- Corte Fino — caixa aberto
INSERT INTO cash_register (id, client_id, opened_by, opened_at, closed_by, closed_at, opening_balance, closing_balance, notes, status) VALUES
('cr-cf-hoje','cli-corte-fino', 'usr-carlos', CURDATE() + INTERVAL 9 HOUR, NULL, NULL, 200.00, NULL, NULL, 'open');

-- ============================================
-- 18. TRANSAÇÕES FINANCEIRAS
-- ============================================
-- Barbearia Modelo — caixa de ontem
INSERT INTO cash_transaction (id, register_id, client_id, appointment_payment_id, type, amount, method, description, category, created_by, created_at) VALUES
('tx-ontem-1', 'cr-ontem', 'cli-barber-modelo', 'pay-h14', 'income',  65.00,  'pix',         'Pgto combo corte+barba — Gustavo',         'Serviços', 'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 9 HOUR + INTERVAL 50 MINUTE),
('tx-ontem-2', 'cr-ontem', 'cli-barber-modelo', 'pay-h15', 'income',  45.00,  'card_credit', 'Pgto corte degradê — Pedro',               'Serviços', 'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 10 HOUR + INTERVAL 30 MINUTE),
('tx-ontem-3', 'cr-ontem', 'cli-barber-modelo', 'pay-h16', 'income',  40.00,  'pix',         'Pgto hidratação — Alexandre',              'Serviços', 'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 14 HOUR + INTERVAL 40 MINUTE),
('tx-ontem-4', 'cr-ontem', 'cli-barber-modelo', NULL,       'income',  89.90,  'cash',        'Venda Pomada Modeladora Matte + Shampoo',  'Vendas',   'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 11 HOUR),
('tx-ontem-5', 'cr-ontem', 'cli-barber-modelo', NULL,       'expense', 30.00,  'cash',        'Compra de café e açúcar',                  'Copa',     'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 8 HOUR + INTERVAL 30 MINUTE),
('tx-ontem-6', 'cr-ontem', 'cli-barber-modelo', NULL,       'income', 210.10,  'card_debit',  'Vendas de produtos diversos no balcão',    'Vendas',   'usr-recep', DATE_SUB(CURDATE(), INTERVAL 1 DAY) + INTERVAL 16 HOUR);

-- Barbearia Modelo — caixa de hoje
INSERT INTO cash_transaction (id, register_id, client_id, appointment_payment_id, type, amount, method, description, category, created_by, created_at) VALUES
('tx-hoje-1', 'cr-hoje', 'cli-barber-modelo', 'pay-t1', 'income',  45.00, 'cash', 'Pgto corte degradê — Lucas', 'Serviços', 'usr-recep', CURDATE() + INTERVAL 9 HOUR + INTERVAL 30 MINUTE),
('tx-hoje-2', 'cr-hoje', 'cli-barber-modelo', NULL,      'expense', 50.00, 'cash', 'Material de limpeza',        'Manutenção','usr-recep', CURDATE() + INTERVAL 8 HOUR + INTERVAL 15 MINUTE),
('tx-hoje-3', 'cr-hoje', 'cli-barber-modelo', NULL,      'income',  59.90, 'pix',  'Venda Pomada Matte no balcão','Vendas',   'usr-recep', CURDATE() + INTERVAL 10 HOUR);

-- ============================================
-- 19. PRODUTOS (ESTOQUE)
-- ============================================
INSERT INTO product (id, client_id, name, description, sku, price, cost_price, quantity_in_stock, low_stock_alert, unit, active) VALUES
('prod-pomada',    'cli-barber-modelo', 'Pomada Modeladora Matte Club',  'Pomada para cabelo efeito fosco e alta fixação',      'PM-MATTE-01',   59.90, 25.00, 14.000, 3.000, 'un', 1),
('prod-shampoo',   'cli-barber-modelo', 'Shampoo Anticaspa Ice 250ml',   'Shampoo refrescante de uso diário',                   'SH-ICE-250',    39.90, 18.00,  2.000, 5.000, 'un', 1),
('prod-condic',    'cli-barber-modelo', 'Condicionador Hidratante 250ml','Condicionador para todos os tipos de cabelo',          'COND-HID-250',  42.90, 19.50, 12.000, 4.000, 'un', 1),
('prod-oleo',      'cli-barber-modelo', 'Óleo para Barba Premium 30ml',  'Óleo hidratante e perfumado para barba',              'OL-BARBA-30',   49.90, 22.00,  0.000, 2.000, 'un', 1),
('prod-goma',      'cli-barber-modelo', 'Goma de Barbear Refresh 100g',  'Facilita o deslizar da lâmina de barbear',            'GOM-BAR-100',   34.90, 15.00,  8.000, 3.000, 'un', 1),
('prod-tonico',    'cli-barber-modelo', 'Tônico Capilar Crescimento',    'Tônico para fortalecimento dos fios',                 'TON-CRESC-50',  89.90, 38.00,  6.000, 2.000, 'un', 1),
('prod-cerveja1',  'cli-barber-modelo', 'Cerveja Heineken Long Neck',    'Cerveja Premium Lager Pilsen',                        'CERV-HEIN-LN',  10.00,  4.50, 24.000,10.000, 'un', 1),
('prod-cerveja2',  'cli-barber-modelo', 'Cerveja Stella Artois LN',      'Cerveja Premium Pilsen Belga',                        'CERV-STELLA-LN',10.00,  4.20, 18.000,10.000, 'un', 1),
('prod-refri',     'cli-barber-modelo', 'Refrigerante Coca-Cola Lata',   'Lata 350ml tradicional',                              'REF-COCA-LATA',  6.00,  2.20, 30.000, 8.000, 'un', 1),
('prod-cafe',      'cli-barber-modelo', 'Café Gourmet Espresso',         'Grãos selecionados moídos na hora',                   'CAF-ESP-01',     5.00,  1.20,  4.500, 1.000, 'kg', 1);

-- ============================================
-- 20. MOVIMENTAÇÕES DE ESTOQUE
-- ============================================
INSERT INTO stock_movement (id, product_id, client_id, type, quantity, reason, appointment_id, created_by, created_at) VALUES
('mov-1', 'prod-pomada',   'cli-barber-modelo', 'in',         20.000, 'Compra lote inicial do fornecedor',         NULL, 'usr-joao', DATE_SUB(NOW(), INTERVAL 7 DAY)),
('mov-2', 'prod-pomada',   'cli-barber-modelo', 'out',         6.000, 'Vendas diretas no balcão',                  NULL, 'usr-recep', DATE_SUB(NOW(), INTERVAL 3 DAY)),
('mov-3', 'prod-shampoo',  'cli-barber-modelo', 'in',         10.000, 'Reposição de estoque',                      NULL, 'usr-joao', DATE_SUB(NOW(), INTERVAL 7 DAY)),
('mov-4', 'prod-shampoo',  'cli-barber-modelo', 'out',         8.000, 'Uso em atendimentos + quebra de frascos',   NULL, 'usr-recep', DATE_SUB(NOW(), INTERVAL 2 DAY)),
('mov-5', 'prod-oleo',     'cli-barber-modelo', 'in',          5.000, 'Compra para revenda',                       NULL, 'usr-joao', DATE_SUB(NOW(), INTERVAL 7 DAY)),
('mov-6', 'prod-oleo',     'cli-barber-modelo', 'adjustment',  0.000, 'Zerado por balanço — todo vendido/consumido',NULL, 'usr-joao', DATE_SUB(NOW(), INTERVAL 1 DAY)),
('mov-7', 'prod-cerveja1', 'cli-barber-modelo', 'in',         48.000, 'Compra mensal de cervejas',                 NULL, 'usr-joao', DATE_SUB(NOW(), INTERVAL 10 DAY)),
('mov-8', 'prod-cerveja1', 'cli-barber-modelo', 'out',        24.000, 'Vendas durante a semana',                   NULL, 'usr-recep', DATE_SUB(NOW(), INTERVAL 2 DAY)),
('mov-9', 'prod-refri',    'cli-barber-modelo', 'in',         36.000, 'Reposição refrigerantes',                   NULL, 'usr-joao', DATE_SUB(NOW(), INTERVAL 5 DAY)),
('mov-10','prod-refri',    'cli-barber-modelo', 'out',         6.000, 'Consumo da semana',                         NULL, 'usr-recep', DATE_SUB(NOW(), INTERVAL 1 DAY));

-- ============================================
-- 21. VÍNCULO SERVIÇO ↔ PRODUTO (consumo automático)
-- ============================================
INSERT INTO service_product (service_id, product_id, client_id, quantity) VALUES
('srv-degrade',    'prod-shampoo', 'cli-barber-modelo', 0.050),
('srv-barba',      'prod-goma',    'cli-barber-modelo', 0.100),
('srv-combo',      'prod-shampoo', 'cli-barber-modelo', 0.050),
('srv-combo',      'prod-goma',    'cli-barber-modelo', 0.080),
('srv-hidratacao', 'prod-condic',  'cli-barber-modelo', 0.200),
('srv-relaxamento','prod-condic',  'cli-barber-modelo', 0.300),
('srv-relaxamento','prod-shampoo', 'cli-barber-modelo', 0.100);

-- ============================================
-- 22. PROGRAMA DE FIDELIDADE
-- ============================================
INSERT INTO loyalty_program (id, client_id, name, type, stamps_to_reward, points_per_real, reward_description, active) VALUES
('lp-modelo', 'cli-barber-modelo', 'Fidelidade Barber Club', 'stamps', 10, NULL, 'Corte de cabelo ou barba grátis ao completar 10 carimbos!', 1);

-- ============================================
-- 23. CARTÕES DE FIDELIDADE DOS CLIENTES
-- ============================================
INSERT INTO loyalty_card (id, customer_id, client_id, program_id, stamps_count, points_balance, status) VALUES
('lcard-pedro',    'cust-pedro',    'cli-barber-modelo', 'lp-modelo', 5,  0.00, 'active'),
('lcard-lucas',    'cust-lucas',    'cli-barber-modelo', 'lp-modelo', 9,  0.00, 'active'),  -- quase premiado!
('lcard-mariana',  'cust-mariana',  'cli-barber-modelo', 'lp-modelo', 1,  0.00, 'active'),
('lcard-carlos',   'cust-carlos',   'cli-barber-modelo', 'lp-modelo', 7,  0.00, 'active'),
('lcard-gabriel',  'cust-gabriel',  'cli-barber-modelo', 'lp-modelo', 3,  0.00, 'active'),
('lcard-bruno',    'cust-bruno',    'cli-barber-modelo', 'lp-modelo', 6,  0.00, 'active'),
('lcard-felipe',   'cust-felipe',   'cli-barber-modelo', 'lp-modelo', 2,  0.00, 'active'),
('lcard-thiago',   'cust-thiago',   'cli-barber-modelo', 'lp-modelo', 4,  0.00, 'active'),
('lcard-matheus',  'cust-matheus',  'cli-barber-modelo', 'lp-modelo', 8,  0.00, 'active');

-- ============================================
-- 24. TRANSAÇÕES DE FIDELIDADE
-- ============================================
INSERT INTO loyalty_transaction (id, card_id, client_id, appointment_id, type, stamps_value, points_value, description, created_by, created_at) VALUES
('ltx-1',  'lcard-pedro',   'cli-barber-modelo', 'apt-h1',  'earn', 1, NULL, 'Carimbo: Corte degradê',          'usr-marcos',  DATE_SUB(NOW(), INTERVAL 5 DAY)),
('ltx-2',  'lcard-pedro',   'cli-barber-modelo', 'apt-h8',  'earn', 1, NULL, 'Carimbo: Retoque degradê',        'usr-marcos',  DATE_SUB(NOW(), INTERVAL 3 DAY)),
('ltx-3',  'lcard-pedro',   'cli-barber-modelo', 'apt-h15', 'earn', 1, NULL, 'Carimbo: Corte degradê rápido',   'usr-tiago',   DATE_SUB(NOW(), INTERVAL 1 DAY)),
('ltx-4',  'lcard-pedro',   'cli-barber-modelo', NULL,       'earn', 2, NULL, 'Carimbos anteriores migrados',    'usr-joao',    DATE_SUB(NOW(), INTERVAL 10 DAY)),
('ltx-5',  'lcard-lucas',   'cli-barber-modelo', 'apt-h2',  'earn', 1, NULL, 'Carimbo: Barba simples',          'usr-tiago',   DATE_SUB(NOW(), INTERVAL 5 DAY)),
('ltx-6',  'lcard-lucas',   'cli-barber-modelo', 'apt-t1',  'earn', 1, NULL, 'Carimbo: Corte degradê hoje',     'usr-marcos',  NOW()),
('ltx-7',  'lcard-lucas',   'cli-barber-modelo', NULL,       'earn', 7, NULL, 'Carimbos anteriores migrados',    'usr-joao',    DATE_SUB(NOW(), INTERVAL 10 DAY)),
('ltx-8',  'lcard-carlos',  'cli-barber-modelo', 'apt-h3',  'earn', 1, NULL, 'Carimbo: Combo corte+barba',      'usr-marcos',  DATE_SUB(NOW(), INTERVAL 5 DAY)),
('ltx-9',  'lcard-carlos',  'cli-barber-modelo', NULL,       'earn', 6, NULL, 'Carimbos anteriores migrados',    'usr-joao',    DATE_SUB(NOW(), INTERVAL 10 DAY)),
('ltx-10', 'lcard-bruno',   'cli-barber-modelo', 'apt-h6',  'earn', 1, NULL, 'Carimbo: Corte degradê',          'usr-marcos',  DATE_SUB(NOW(), INTERVAL 4 DAY)),
('ltx-11', 'lcard-bruno',   'cli-barber-modelo', NULL,       'earn', 5, NULL, 'Carimbos anteriores migrados',    'usr-joao',    DATE_SUB(NOW(), INTERVAL 10 DAY)),
('ltx-12', 'lcard-matheus', 'cli-barber-modelo', 'apt-h9',  'earn', 1, NULL, 'Carimbo: Combo corte+barba',      'usr-tiago',   DATE_SUB(NOW(), INTERVAL 3 DAY)),
('ltx-13', 'lcard-matheus', 'cli-barber-modelo', NULL,       'earn', 7, NULL, 'Carimbos anteriores migrados',    'usr-joao',    DATE_SUB(NOW(), INTERVAL 10 DAY));

-- ============================================
-- PRONTO! SEED COMPLETO EXECUTADO.
-- ============================================
