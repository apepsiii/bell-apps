-- Seeder: Data from SQLite database.db
-- Run after migration 000002

-- majors (5 rows)
INSERT INTO majors (id, name) VALUES (2, 'Otomatisasi dan Tata Kelola Perkantoran');
INSERT INTO majors (id, name) VALUES (3, 'Manajemen Perkantoan dan Layanan Bisnis');
INSERT INTO majors (id, name) VALUES (4, 'Pemasaran');
INSERT INTO majors (id, name) VALUES (5, 'Bisnis Daring dan Pemasaran');
INSERT INTO majors (id, name) VALUES (6, 'ALUMNI');

-- classes (5 rows)
INSERT INTO classes (id, name, major_id, wa_group_id) VALUES (14, 'ALUMNI', 6, '');
INSERT INTO classes (id, name, major_id, wa_group_id) VALUES (15, 'XII-OTKP', 2, '');
INSERT INTO classes (id, name, major_id, wa_group_id) VALUES (16, 'XII-BDP', 5, '');
INSERT INTO classes (id, name, major_id, wa_group_id) VALUES (17, 'XI-MPLB', 3, '');
INSERT INTO classes (id, name, major_id, wa_group_id) VALUES (18, 'XI-PM', 4, '');

-- students (156 rows)
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (1, '2515254803', '2515254803', 'ABDULAH FATHUR RAHMAN ASSYAHIDIN', '6285811158295', 'SITI MUAWANAH', 14, '2515254803_20260208183855.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (2, '2513121059', '2513121059', 'ADISTY RAHMANDA PUTRI', '629512650988', 'Iman Supriatman', 14, '2513121059_20260419155320.jpg', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (3, '2512307523', '2512307523', 'ALYSHA DWI VANIA', '6289610742595', 'ROSITA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (4, '2511570915', '2511570915', 'AMANDA RIZYA AKBAR', '62895320056561', 'RITA LESTARI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (5, '2512999395', '2512999395', 'ANZILNA NAURAL MUMTAZAH', '628567955557', 'LAILYA MAFLUKHAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (6, '4267654589', '4267654589', 'DANIEL KRISTOFEL ENRISSON HASIBUAN', '628176996456', 'HENNY P.C.SINAGA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (7, '2513691955', '2513691955', 'ELPANI ROSDIANTI', '6285863043357', 'NENENG', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (8, '2511354419', '2511354419', 'FADHILLAH', '6283872828514', 'Daelami', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (9, '552965605', '552965605', 'HELDES', '6285930269041', 'Eni', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (10, '2971005792', '2971005792', 'JELITA APRILIA PUTRI', '6289541597079', 'Yeyet Nuryati', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (11, '549119525', '549119525', 'KEYLA AHMAD', '6283849546470', 'Devi Susriani', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (12, '2509419331', '2509419331', 'MAHDIYAH ZAHRA', '6287888188208', 'Ayu karmila', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (13, '2513055987', '2513055987', 'MAULANA YUSUF IBRAHIM', '62895375729893', 'DEDEH KURNIASIH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (14, '2521384403', '2521384403', 'MOCHAMAD CIPTA ANUGRAH', '628316073882', 'Citra asriana', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (15, '2509411267', '2509411267', 'MOHAMAD JOVAN ERLANGGA', '6283133967799', 'EVA ROSEDIANA DEWI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (16, '2514532035', '2514532035', 'MUHAMAD ADITYA ALFARIZI', '6285210158327', 'MARISAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (17, '2511763507', '2511763507', 'MUHAMAD IQBAL ALFARIZ', '0851', 'SRI SUHARTINI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (18, '222310313', '222310313', 'HELENA SAFRIANTI', '08132819021891', 'Hellesi Adnami', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (19, '2516596547', '2516596547', 'MUHAMMAD AZHAR NAWAWI', '6285692025754', 'SUMIRAT', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (20, '2509700419', '2509700419', 'NENG TESSA', '62838133442258', 'Ajam', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (21, '2512157667', '2512157667', 'NURAINISA POETRI', '6285890296672', 'SUKAESIH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (22, '2519069907', '2519069907', 'NURLAILIA', '6283848490149', 'MIMAH SAHIMAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (23, '2513903059', '2513903059', 'RAHELIA', '6285892930291', 'RATNA SUMINAR', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (24, '2514949587', '2514949587', 'RESA AGUSTINA', '6283895457028', 'NUNUNG', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (25, '2513462211', '2513462211', 'RIANTI ARDIANA PUTRI', '6285694470881', 'Adung syahrudin', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (26, '2970854784', '2970854784', 'SELVI AGNIA PUTRI', '6281310783701', 'LENI MARLINA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (27, '2511701523', '2511701523', 'SITI ALISA', '6283105832057', 'SITI NURIPAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (28, '2513130291', '2513130291', 'SITI KURNIA SARI', '6285779352519', 'Luna silvia', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (29, '2517868563', '2517868563', 'SITI NURMAYANTI MAULIDINI', '6283877379811', 'Ipah', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (30, '2514576547', '2514576547', 'SRI HARYATI', '6289657221282', 'ATI ATIKAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (31, '2518782147', '2518782147', 'SYALMAN ALFARIZI', '6281315194533', 'Suhendar', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (32, '2507632947', '2507632947', 'ABIMAYU AURORA MARANTIKA', '6281310783701', 'IBU EMI SUHAEMI', 14, '2507632947_20260208183906.jpg', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (33, '222310301', '222310301', 'ADINDA SARAH GUMATI', '6281310783701', 'Sariwati', 14, '222310301_20260208190256.jpg', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (34, '2510685379', '2510685379', 'ANDINI YULISTIANINGSIH', '08985277370', 'IBU ICE IRIANAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (35, '2512144883', '2512144883', 'AZKA AULIA', '6285210158081', 'IBU SUMIATI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (36, '2512952739', '2512952739', 'CYNDI GELANTIA STEFANIE', '62895388817700', 'IBU ANGGI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (37, '2507550627', '2507550627', 'DEA SALFA RAHMA', '6282142086157', 'IBU NILA KUSUMA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (38, '2511952691', '2511952691', 'DIFFA SETIA SANDRA KIRANA', '6289523008269', 'IBU SULASTRI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (39, '2513586099', '2513586099', 'EFVRYLLIA SALSABILA', '6289523008269', 'IBU FITRI YANI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (40, '2510903827', '2510903827', 'ERGI AULIANDRA', '6285875400478', 'IBU SARI NURJANAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (41, '2513638979', '2513638979', 'ERINA DAANYA AMELIA', '6285212631213', 'AANG HAYATI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (42, '2507439443', '2507439443', 'FAIZ FATHUL ISLAM', '6289518751979', 'IBU TIENI ISWARNI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (43, '2514599763', '2514599763', 'GAGAS SUPRIATNA PUTRA', '6283183638615', 'IBU EEN YULIA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (44, '2515224611', '2515224611', 'M. ABDUL KARIM DZULHIAGA', '6281574275798', 'IBU WITA ANISA SEPTIAGA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (45, '2510755011', '2510755011', 'MOHAMAD VIZRI ZAELANI', '6283893159472', 'IBU NURAENI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (46, '2511805427', '2511805427', 'MUHAMAD ARIF RAMDANI', '6282310767760', 'IBU HANNA MUNTAHANAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (47, '2509995667', '2509995667', 'MUHAMAD IKBAL NUR IMAN', '6285780985052', 'IBU TITIN PATIMAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (48, '2511394547', '2511394547', 'MUHAMAD RAIHAN FARISAL', '6289517445595', 'IBU SUSI SULISTIANAWATY', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (49, '2511398803', '2511398803', 'MUHAMAD SAEPUL RIZKI', '62895320463532', 'IBU SITI JENAB', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (50, '2513899635', '2513899635', 'NABILA NADIA ERNAS', '62895385841243', 'IBU ERNI MARYANI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (51, '2513272227', '2513272227', 'NABILA ZALWA OKTAVIANI', '6282124900801', 'IBU ARYANTI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (52, '552870533', '552870533', 'NAZWA NAULIA PUTRI', '6285748752381', 'IBU ITA ROSITA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (53, '2511356867', '2511356867', 'PUTRI MARYANTI', '6283872822039', 'IBU UNAR JUNIARSIH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (54, '2507533827', '2507533827', 'SALMA DESFANIA AZ ZAHRA', '6289663611012', 'IBU WAWAT HERAWATI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (55, '2519210851', '2519210851', 'SELA JULIANTI', '6281293755188', 'JUBAEDAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (56, '2513414755', '2513414755', 'SITI FANIYAH', '6285894861643', 'IBU ASIH HERNAWATI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (57, '2511552227', '2511552227', 'SYIFA AULIA', '6281287916620', 'IBU DEDEH KURNIASIH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (58, '2514577923', '2514577923', 'TEGUH IRWANSYAH', '6281997317797', 'BAPAK WAWAN GUNAWAN', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (59, '2509894387', '2509894387', 'TIA NURDIANTI', '6281401786380', 'IBU DEDE HERLIAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (60, '4267792717', '4267792717', 'ADNAN MUHAMMAD ABDULLAH', '6285890198663', 'BADRIAH', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (61, '2970990624', '2970990624', 'AMANDA DINAR', '6283879736504', 'CHRISTINE NATALIA', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (62, '2971077488', '2971077488', 'ELVIRA AZZAHRA', '6289521070418', 'DEDE NANI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (63, '4268865693', '4268865693', 'LAILA RAMADANI', '6285175429180', 'JULAEHA', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (64, '4267571757', '4267571757', 'LINA ANGGRAENI', '6287820783587', 'TUKKIMIN DIAN WAHYUNI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (65, '4268764973', '4268764973', 'MUHAMAD RAKHA SETIAWAN', '62895423035325', 'FIKKI SETIAWAN', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (66, '4267936877', '4267936877', 'MUHAMMAD RIJAL', '62895355313754', 'ICAH WAWATI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (67, '4267839821', '4267839821', 'MUHAMMAD RIZKY PENWAR', '62895320083453', 'ABDUL KADIR', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (68, '4267630909', '4267630909', 'MUHAMMAD SYAHRUL PRATAMA', '628976390560', 'INDAH SUSANTI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (69, '4268516381', '4268516381', 'NUR FAIDAH', '6287830193347', 'YATI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (70, '4267665597', '4267665597', 'PUTRI MELATI', '6289638271413', 'FAJAR', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (71, '2973541296', '2973541296', 'RISKA CAHYA ANJANI', '628979258677', 'SITI WAHIDAH', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (72, '222310342', '222310342', 'HESTI SURAYA NINGSIH', '6281912907234', 'HILMI HESTIAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (73, '2956973696', '2956973696', 'SARAH AMELIA', '6289513261872', 'SARIPUDIN', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (74, '2970698832', '2970698832', 'SEPTI RAMADHANI', '62895349003165', 'YAYU RAHAYU', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (75, '2970661552', '2970661552', 'SERLI CAHYA NURDINI', '6285894105820', 'NINING', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (76, '2970906832', '2970906832', 'SITI FATIMAH AZ-ZAHRA', '6285780259547', 'UJANG SOPARI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (77, '2956391200', '2956391200', 'SITI HABIBAH', '6285714918679', 'SITI JUARSIH', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (78, '2970746720', '2970746720', 'SITI LUTFIAH NUR SAWILYAH', '6285693553785', 'MARYANI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (79, '2971273952', '2971273952', 'SITI NUR THALITA', '6283133327899', 'KIAH', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (80, '2971266224', '2971266224', 'SITI SALMA AZKIA', '6289654353115', 'RIKA SUSILAWATI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (81, '2970984864', '2970984864', 'SYALWA AS SAEFU ROMADONA', '6283840901207', 'SAEPUDIN', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (82, '2971134512', '2971134512', 'TEDDY RAHMAT SUPRIATNA', '6283815553204', 'ALIYAH', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (83, '2971112192', '2971112192', 'VEGA AULIA', '62851710864382', 'NENENG', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (84, '4267729837', '4267729837', 'ZAHRA KAMILA WIJAYA', '6285814137273', 'JAYA', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (85, '4267892717', '4267892717', 'ZAHRA ZASKI AULIA', '6283813586436', 'ADE SUGANDA', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (86, '4267796637', '4267796637', 'AFRIL NUR ASSYIFA', '6285710744721', 'ATIN SOFYAN', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (87, '2970675168', '2970675168', 'AGISTY SULISTIYU', '6285717033320', 'YULI AGUSTINI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (88, '4267701533', '4267701533', 'AJENG AYUNI PUTRI', '6283811198359', 'YENI ROYANI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (89, '2970977792', '2970977792', 'ALVIANSYAH PRIASETYA NAZAR', '6285888315340', 'ALVIN LEE NAZAR', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (90, '2970861920', '2970861920', 'ANANDA ASRI RAMADANTI', '6285694801992', 'RINI KARTINI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (91, '2970371472', '2970371472', 'ANNISA NUR FEBRIYANTY', '6285888414432', 'ETI ROHATI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (92, '2970365152', '2970365152', 'AZKIA NURUL FAJARI', '6285811933703', 'ANITA', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (93, '2970994384', '2970994384', 'CLARA ANNA AGUSTINA', '6282111572322', 'YANI SUMIYATI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (94, '4268612013', '4268612013', 'DEASI CAHYANI', '6289611498290', 'SUPRIYATTIH', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (95, '2970284752', '2970284752', 'DESWA ADHANI', '6288975917869', 'YULIANTI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (96, '2970925024', '2970925024', 'FAKHRI DWIANDRA MAULANDI', '6289518394956', 'MAMUNG', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (97, '549167989', '549167989', 'HESTINAH', '6283876901016', 'IPAH', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (98, '2970943216', '2970943216', 'INDAH MAYANG SARI', '6285710654610', 'LILIS MARYANI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (99, '4267839837', '4267839837', 'MUHAMAD ABIDIN', '6281389916982', 'OMA', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (100, '4267662685', '4267662685', 'MUHAMAD IHSAN', '6283811296004', 'ICHWAN SUWANDI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (101, '4267718557', '4267718557', 'MUHAMAD NUR WAHYU', '6285445958020', 'NUR', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (102, '552846421', '552846421', 'MUTIA SABILA', '6283856440394', 'SAPPUDIN', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (103, '4267911501', '4267911501', 'MOH. VICKY FIRANSYA', '6285718647528', 'WAWAN KURNIAWAN', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (104, '4267716509', '4267716509', 'NATASYA PUTRI', '6285718105141', 'NATA SASMITA', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (105, '4268593949', '4268593949', 'NAYSYILA HERAWATI', '6283894965160', 'HURAINI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (106, '4267749965', '4267749965', 'PUTRI FAIHA AL - ROJAK', '628381866606', 'JERI OJAK SETIAWAN', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (107, '2971126944', '2971126944', 'SITI NUR FADILLA', '6282311448345', 'M.HERMAWAN', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (108, '4268575341', '4268575341', 'TRISA APRILIA PUTRI', '629603960429', 'SANTI ELAWATI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (109, '2970652432', '2970652432', 'VISKA CANTIKA', '6285711903628', 'ALAM SARI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (110, '2970888032', '2970888032', 'YUDISTI ZETIRA RAMADAN', '6285719340689', 'SRI SUMASTINI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (111, '4267849741', '4267849741', 'ZAHRA ABDULLAH', '6283189962174', 'MARIANAH', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (112, '2959012784', '2959012784', 'ZAQI MUHAMAD AKBAR', '6281288463741', 'IMAS', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (113, '4268751693', '4268751693', 'ZIDAN MAULIDAN', '6285710308629', 'RIANI HANDAYANI', 16, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (114, '3730162260', '3730162260', 'ADJMA NURPIRDA', '6285891695321', 'IBU ERNI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (115, '3729794356', '3729794356', 'AGIESKHA VIANA SIAMY', '6285281021740', 'IBU NENENG', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (116, '550791109', '550791109', 'ALISA APRIL YANTI', '6289694866668', 'MOHTAR', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (117, '553069077', '553069077', 'ANGGYA ESA MAULIDA', '6285693813207', 'SUPIAH', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (118, '552675301', '552675301', 'DEA CITRA APRILIA', '6289505746755', 'ANI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (119, '553089493', '553089493', 'DENIS MUHAMAD RISQI', '628995768152', 'LILIM', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (120, '549171909', '549171909', 'MAUDI ANUGRAH HEKSA PUTRI', '6281802047662', 'IBU ENTIN', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (121, '549191557', '549191557', 'MELLA ROSE', '6281513915893', 'IBU MERRY', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (122, '548948613', '548948613', 'NURI MAULIDA', '6283194413613', 'IBU SITI HASANAH', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (123, '550443845', '550443845', 'SASKIA AQILA KAADZIYAH', '6281380268110', 'IBU SITI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (124, '551111301', '551111301', 'SHIFFA APRILIA', '6283808767732', 'IBU DESI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (125, '551580789', '551580789', 'SITI ASTRI APSIAH', '6289503138875', 'IBU RIKA', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (126, '551226949', '551226949', 'SITI DELA DAVINA RAMADHANI', '6283192632433', 'BAPAK DANI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (127, '551168133', '551168133', 'TASSYA NURMAENI', '6285770194977', 'IBU ENI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (128, '551371365', '551371365', 'TAUFIK RAU''UF ERYADI', '62895355354174', 'IBU RAHAYU', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (129, '548084453', '548084453', 'UMAR SUBAGYA SUBARKAH', '6285714052828', 'ANDI SUGANDI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (130, '3772860212', '3772860212', 'WINDI MEIDI', '6281399791064', 'VINA OKTAVIANI', 17, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (131, '549360453', '549360453', 'AZAHRA ADELIA PUTRI', '0887', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (132, '551072933', '551072933', 'DELLIA OCTHAVIANI', '0888', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (133, '552602565', '552602565', 'DESWITA PUTRI ANGGRAENI', '0889', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (134, '552640965', '552640965', 'FITRI SETIAWAN', '0890', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (135, '553042661', '553042661', 'INGGRIT NURGRITA SARI', '0891', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (136, '552868901', '552868901', 'INKA SINTYA MAYLAIKA', '0892', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (137, '548986965', '548986965', 'KEVIN ARYANSYAH', '0893', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (138, '549192053', '549192053', 'MAYA DEFRIANTI', '0894', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (139, '548702405', '548702405', 'MUTIARA HANDAYANI', '0895', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (140, '551578325', '551578325', 'NOURI NASHIRA', '0896', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (141, '548539717', '548539717', 'SALSABILA', '0897', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (142, '548678293', '548678293', 'SANDYRA CENDANA JATI. P', '0898', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (143, '549291093', '549291093', 'SEPTIAN RISKI', '0899', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (144, '551176181', '551176181', 'SHILVA RAMADHANI', '0900', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (145, '551251733', '551251733', 'SITI JULIANISA PUTRI', '0901', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (146, '550747285', '550747285', 'SITI MELATI', '0902', 'NAMA WALI', 18, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (147, '2514604099', '2514604099', 'SITI ZAHRA MUSLIANI', '628121', 'Nama Wali', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (148, '548539541', '548539541', 'RASHIED UMAR RAZIEB', '6285888666720', 'DEDE HERLIAH', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (149, '549187941', '549187941', 'M. HILMAN MAULANA', '6285881036291', 'S. SANJAYA', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (150, '548273589', '548273589', 'FEBIYANTI', '6289670099163', 'JURIAH', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (151, '2510886675', '2510886675', 'RAHMAH NURMILAH SARI', '62813090291', 'Wali', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (152, '222310314', '222310314', 'IRFAN', '0812121212', 'IRNAWATI', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (153, '222310358', '222310358', 'RIFA RASIYA', '628102819378', 'RIANTI SYARIA', 14, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (154, '232410376', '232410376', 'MEYSIE AYU ANGGRIANI', '', 'REYGI ABDULFAJAR', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (155, '232410388', '232410388', 'SHIRREN APRILIA WAHYUDI', '', 'SARI SUNARDI', 15, 'nopic.png', '');
INSERT INTO students (id, rfid_uid, nis, name, parent_phone, parent_name, class_id, photo, birthday) VALUES (309, '232410375', '232410375', 'MARSYA DWI PERMANA', '', 'SITI SARAH', 15, 'nopic.png', '');

-- staff (1 rows)
INSERT INTO staff (id, rfid_uid, nip, name, phone, role) VALUES (1, '11', '112', 'Ape', '6281310783701', 'Guru');

-- operators (1 rows)
INSERT INTO operators (id, username, password, name, phone, photo, is_active, created_at) VALUES (1, 'operator', '$2a$10$I24momJeERixDtBC4kuzIO7YcshOYI1IgpeItHuN2D62IPMVVB3Oq', 'Operator Default', '081234567890', NULL, 1, '2026-04-19 07:40:49');

-- schedules (2 rows)
INSERT INTO schedules (id, time, label, audio_file) VALUES (2, '07:00', 'Jam Pelajaran Pertama', 'Pelajaran_ke_1_20251111_100632.wav');
INSERT INTO schedules (id, time, label, audio_file) VALUES (3, '23:40', 'Testing', 'Pelajaran_ke_1_20251111_100632.wav');

-- audio_files (6 rows)
INSERT INTO audio_files (id, file_name, display_name) VALUES (3, '5_Menit_Pelajaran_ke_1_20251111_100557.wav', '5_Menit_Pelajaran_ke_1_20251111_100557.wav');
INSERT INTO audio_files (id, file_name, display_name) VALUES (4, 'Pelajaran_ke_1_20251111_100632.wav', 'Pelajaran_ke_1_20251111_100632.wav');
INSERT INTO audio_files (id, file_name, display_name) VALUES (5, 'Akhir_Pelajaran_20251111_100621.wav', 'Akhir_Pelajaran_20251111_100621.wav');
INSERT INTO audio_files (id, file_name, display_name) VALUES (6, 'Bangun Pemudi Pemuda - Lirik Lagu Nasional Indonesia.mp3', 'Bangun Pemudi Pemuda - Lirik Lagu Nasional Indonesia.mp3');
INSERT INTO audio_files (id, file_name, display_name) VALUES (7, 'Asmaul Husna - Lagu 99 Nama Allah yang Merdu.mp3', 'Asmaul Husna - Lagu 99 Nama Allah yang Merdu.mp3');
INSERT INTO audio_files (id, file_name, display_name) VALUES (8, 'Arab_Jam_1_Mulai_20251111_011535.wav', 'Arab_Jam_1_Mulai_20251111_011535.wav');

-- devices (1 rows)
INSERT INTO devices (id, name, ip_address, status, last_sync) VALUES (3, 'Front Office', '192.168.1.1', 'offline', '-');

-- holidays (7 rows)
INSERT INTO holidays (id, date, name, type, description, created_at) VALUES (1, '2026-06-01', 'Hari Lahir Pancasila', 'national', 'Libur Nasional Standar', '2026-02-04 16:41:31');
INSERT INTO holidays (id, date, name, type, description, created_at) VALUES (2, '2026-08-17', 'Hari Kemerdekaan RI', 'national', 'Libur Nasional Standar', '2026-02-04 16:41:31');
INSERT INTO holidays (id, date, name, type, description, created_at) VALUES (3, '2026-12-25', 'Hari Raya Natal', 'national', 'Libur Nasional Standar', '2026-02-04 16:41:32');
INSERT INTO holidays (id, date, name, type, description, created_at) VALUES (4, '2026-01-01', 'Tahun Baru Masehi', 'national', 'Libur Nasional Standar', '2026-02-04 16:41:32');
INSERT INTO holidays (id, date, name, type, description, created_at) VALUES (5, '2026-05-01', 'Hari Buruh Internasional', 'national', 'Libur Nasional Standar', '2026-02-04 16:41:32');
INSERT INTO holidays (id, date, name, type, description, created_at) VALUES (9, '2026-02-17', 'Libur Imlek', 'national', 'Imlek', '2026-02-08 15:33:07');
INSERT INTO holidays (id, date, name, type, description, created_at) VALUES (10, '2026-02-17', 'Libur Imlek', 'national', 'Imlek', '2026-02-08 15:33:07');

-- school_settings (1 rows)
INSERT INTO school_settings (id, setting_key, setting_value) VALUES (1, 'work_days', '1,2,3,4,5');

-- point_rules (2 rows)
INSERT INTO point_rules (id, category, name, points, description) VALUES (1, 'achievement', 'AAA', 23, 'sa');
INSERT INTO point_rules (id, category, name, points, description) VALUES (2, 'violation', 'asasax', -20, 'adas');

-- point_rewards (1 rows)
INSERT INTO point_rewards (id, name, points_cost, stock, description) VALUES (1, 'KS', 5000, 12, 'asasa');

-- student_points (1 rows)
INSERT INTO student_points (id, student_id, rule_id, reward_id, points_change, description, timestamp, recorded_by) VALUES (1, 1, 1, NULL, 23, 'AAA', '2026-02-08 17:54:47', 'Admin');

-- attendance_settings (21 rows)
INSERT INTO attendance_settings (setting_key, setting_value) VALUES
  ('wa_template_staff', '✅ KONFIRMASI PRESENSI GURU/STAF

Yth. Bapak/Ibu {teacher_name},

Presensi {type} Anda pada hari {date} telah berhasil dicatat sistem.

🕒 Pukul: {time} WIB

[Jika tipe Masuk] Selamat bertugas dan semoga hari Anda menyenangkan! [Jika tipe Pulang] Terima kasih atas dedikasi hari ini. Selamat beristirahat.

— Sistem Presensi Sekolah —'),
  ('arrival_start', '00:00'),
  ('arrival_end', '00:30'),
  ('departure_start', '15:30'),
  ('departure_end', '21:00'),
  ('onesender_api_url', 'https://sender3.pionireradigital.id/api/v1/message'),
  ('onesender_api_token', 'uf98a57a24342443.10519e46931d42498b6bf326fac524c2'),
  ('wa_template_in', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 {name}

Telah melakukan presensi pada: 
🕒 Pukul: {time} 
✅ Status: {status}

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.'),
  ('wa_template_late', '⚠️ INFO KETERLAMBATAN SISWA

Yth. Bapak/Ibu Orang Tua/Wali,

Kami informasikan bahwa putra/putri Anda: 
👤 Nama: {name} 
📅 Tanggal: {date} 
⏰ Tiba Pukul: {time} WIB

Status kehadiran hari ini tercatat: *TERLAMBAT.*

Siswa telah diizinkan masuk kelas untuk mengikuti pelajaran. Mohon kerjasama Bapak/Ibu untuk mendorong kedatangan tepat waktu di kemudian hari.

Terima kasih. 
_— Sistem Presensi Sekolah —_'),
  ('wa_template_out', '🏠 INFO KEPULANGAN SISWA

Yth. Orang Tua/Wali {name},

Kami informasikan bahwa kegiatan belajar mengajar hari ini, {date}, telah selesai.

Putra/putri Anda telah melakukan presensi pulang pada pukul {time} WIB dan telah meninggalkan area sekolah.

Mohon dipantau kedatangannya di rumah. Terima kasih.

— Sistem Presensi Sekolah —'),
  ('wa_template_staff_in', '_Assalamualaikum Warahmatullahi Wabarakaatuh_

☀️ SELAMAT PAGI, PAK/BU GURU!

Yth. {teacher_name},

Presensi {type} Anda pada tanggal {date} telah berhasil dicatat oleh sistem.

🕒 Pukul: {time} 
📋 Status: {status}

Selamat menjalankan tugas mulia mencerdaskan kehidupan bangsa. Semoga hari Anda menyenangkan! ✨

— Sistem Kepegawaian Sekolah —'),
  ('wa_template_staff_out', '_Assalamualaikum Warahmatullahi Wabarakaatuh_

🌙 TERIMA KASIH ATAS DEDIKASINYA

Yth. Bapak/Ibu {teacher_name},

Kegiatan sekolah hari ini, {date}, telah usai. Data presensi Anda telah kami simpan:

🔄 Tipe: {type} 
🕒 Pukul: {time} 
✅ Validasi: Sukses

Selamat beristirahat dan sampai jumpa kembali esok hari. Hati-hati di jalan! 🙏

— Sistem Kepegawaian Sekolah —'),
  ('wa_image_link', 'https://services.presensi.co.id/media/logometa.png'),
  ('dzuhur_start', '11:30'),
  ('dzuhur_end', '13:00'),
  ('ashar_start', '15:00'),
  ('ashar_end', '16:00'),
  ('birthday_enabled', 'true'),
  ('wa_template_birthday', ''),
  ('wa_image_birthday', 'https://cdn.bakedbree.com/uploads/2024/07/Chocolate-Birthday-Cake-A_chocolate_birthday_cake_feature_4.jpg'),
  ('birthday_time', '16:30');

-- attendance_logs (226 rows)
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (1, '1', 'Testing ah', 'Siswa', 'Datang', 'RFID', '2026-01-26 00:06:32', '2026-01-26');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (2, '2', 'Asa', 'Siswa', 'Sakit', 'RFID', '2026-01-26 00:11:21', '2026-01-26');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (3, '3', 'Apepsiii', 'Siswa', 'Datang', 'RFID', '2026-01-26 00:27:17', '2026-01-26');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (4, '4', 'Testing ah asa', 'Siswa', 'Terlambat', 'RFID', '2026-01-26 00:43:06', '2026-01-26');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (5, '11', 'Ape', 'Staff', 'Terlambat', 'RFID', '2026-01-26 01:00:23', '2026-01-26');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (6, '3', 'Apepsiii', 'Siswa', 'Datang', 'MANUAL', '2026-02-04 21:37:15', '2026-02-04');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (7, '1', 'Testing ah', 'Siswa', 'Sakit', 'MANUAL', '2026-02-04 21:37:19', '2026-02-04');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (8, '4', 'Testing ah asa', 'Siswa', 'Datang', 'MANUAL', '2026-02-04 21:37:23', '2026-02-04');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (9, '2', 'Asa', 'Siswa', 'Alpha', 'MANUAL', '2026-02-04 21:37:27', '2026-02-04');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (10, '3', 'Apepsiii', 'Siswa', 'Datang', 'MANUAL', '2026-02-05 12:00:25', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (11, '2', 'Asa', 'Siswa', 'Datang', 'MANUAL', '2026-02-05 12:00:35', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (12, '1', 'Testing ah', 'Siswa', 'Sakit', 'MANUAL', '2026-02-05 12:00:44', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (13, '4', 'Testing ah asa', 'Siswa', 'Alpha', 'MANUAL', '2026-02-05 15:25:15', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (14, '3', 'Apepsiii', 'Siswa', 'Datang', 'MANUAL', '2026-02-06 11:48:46', '2026-02-06');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (15, '1', 'Testing ah', 'Siswa', 'Terlambat', 'RFID', '2026-02-06 15:17:01', '2026-02-06');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (16, '2', 'Asa', 'Siswa', 'Terlambat', 'RFID', '2026-02-06 15:17:20', '2026-02-06');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (17, '4', 'Testing ah asa', 'Siswa', 'Terlambat', 'RFID', '2026-02-06 15:17:23', '2026-02-06');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (18, '1', 'Testing ah', 'Siswa', 'Pulang', 'RFID', '2026-02-06 15:39:33', '2026-02-06');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (19, '2', 'Asa', 'Siswa', 'Pulang', 'RFID', '2026-02-06 15:39:38', '2026-02-06');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (20, '4', 'Testing ah asa', 'Siswa', 'Pulang', 'RFID', '2026-02-06 15:48:25', '2026-02-06');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (21, '2515254803', 'ABDULAH FATHUR RAHMAN ASSYAHIDIN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 19:06:09', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (23, '2507632947', 'ABIMAYU AURORA MARANTIKA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 20:15:34', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (24, '222310301', 'ADINDA SARAH GUMATI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 21:42:10', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (25, '2513121059', 'ADISTY RAHMANDA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (26, '2512307523', 'ALYSHA DWI VANIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (27, '2511570915', 'AMANDA RIZYA AKBAR', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (28, '2510685379', 'ANDINI YULISTIANINGSIH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (29, '2512999395', 'ANZILNA NAURAL MUMTAZAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (30, '2512144883', 'AZKA AULIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (31, '2512952739', 'CYNDI GELANTIA STEFANIE', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (32, '4267654589', 'DANIEL KRISTOFEL ENRISSON HASIBUAN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (33, '2507550627', 'DEA SALFA RAHMA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (34, '2511952691', 'DIFFA SETIA SANDRA KIRANA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (35, '2513586099', 'EFVRYLLIA SALSABILA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (36, '2513691955', 'ELPANI ROSDIANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (37, '2510903827', 'ERGI AULIANDRA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (38, '2513638979', 'ERINA DAANYA AMELIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (39, '2511354419', 'FADHILLAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (40, '2507439443', 'FAIZ FATHUL ISLAM', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (41, '2514599763', 'GAGAS SUPRIATNA PUTRA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (42, '552965605', 'HELDES', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (43, '222310313', 'HELENA SAFRIANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (44, '222310342', 'HESTI SURAYA NINGSIH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (45, '222310314', 'IRFAN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (46, '2971005792', 'JELITA APRILIA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (47, '549119525', 'KEYLA AHMAD', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (48, '2515224611', 'M. ABDUL KARIM DZULHIAGA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (49, '2509419331', 'MAHDIYAH ZAHRA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (50, '2513055987', 'MAULANA YUSUF IBRAHIM', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (51, '2521384403', 'MOCHAMAD CIPTA ANUGRAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (52, '2509411267', 'MOHAMAD JOVAN ERLANGGA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (53, '2510755011', 'MOHAMAD VIZRI ZAELANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (54, '2514532035', 'MUHAMAD ADITYA ALFARIZI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (55, '2511805427', 'MUHAMAD ARIF RAMDANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (56, '2509995667', 'MUHAMAD IKBAL NUR IMAN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (57, '2511763507', 'MUHAMAD IQBAL ALFARIZ', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (58, '2511394547', 'MUHAMAD RAIHAN FARISAL', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (59, '2511398803', 'MUHAMAD SAEPUL RIZKI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (60, '2516596547', 'MUHAMMAD AZHAR NAWAWI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (61, '2513899635', 'NABILA NADIA ERNAS', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (62, '2513272227', 'NABILA ZALWA OKTAVIANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (63, '552870533', 'NAZWA NAULIA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (64, '2509700419', 'NENG TESSA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (65, '2512157667', 'NURAINISA POETRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (66, '2519069907', 'NURLAILIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (67, '2511356867', 'PUTRI MARYANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (68, '2513903059', 'RAHELIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (69, '2510886675', 'RAHMAH NURMILAH SARI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (70, '548539541', 'RASHIED UMAR RAZIEB', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (71, '2514949587', 'RESA AGUSTINA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (72, '2513462211', 'RIANTI ARDIANA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (73, '222310358', 'RIFA RASIYA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (74, '2507533827', 'SALMA DESFANIA AZ ZAHRA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (75, '2519210851', 'SELA JULIANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (76, '2970854784', 'SELVI AGNIA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (77, '2511701523', 'SITI ALISA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (78, '2513414755', 'SITI FANIYAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (79, '2513130291', 'SITI KURNIA SARI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (80, '2517868563', 'SITI NURMAYANTI MAULIDINI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (81, '2514604099', 'SITI ZAHRA MUSLIANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (82, '2514576547', 'SRI HARYATI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (83, '2518782147', 'SYALMAN ALFARIZI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (84, '2511552227', 'SYIFA AULIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (85, '2514577923', 'TEGUH IRWANSYAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (86, '2509894387', 'TIA NURDIANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (87, '3730162260', 'ADJMA NURPIRDA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (88, '3729794356', 'AGIESKHA VIANA SIAMY', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (89, '550791109', 'ALISA APRIL YANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (90, '553069077', 'ANGGYA ESA MAULIDA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (91, '552675301', 'DEA CITRA APRILIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (92, '553089493', 'DENIS MUHAMAD RISQI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (93, '549171909', 'MAUDI ANUGRAH HEKSA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (94, '549191557', 'MELLA ROSE', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (95, '548948613', 'NURI MAULIDA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (96, '550443845', 'SASKIA AQILA KAADZIYAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (97, '551111301', 'SHIFFA APRILIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (98, '551580789', 'SITI ASTRI APSIAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (99, '551226949', 'SITI DELA DAVINA RAMADHANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (100, '551168133', 'TASSYA NURMAENI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (101, '551371365', 'TAUFIK RAU''UF ERYADI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (102, '548084453', 'UMAR SUBAGYA SUBARKAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (103, '3772860212', 'WINDI MEIDI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (104, '4267792717', 'ADNAN MUHAMMAD ABDULLAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 21:45:17', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (105, '3730162260', 'ADJMA NURPIRDA', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (106, '3729794356', 'AGIESKHA VIANA SIAMY', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (107, '550791109', 'ALISA APRIL YANTI', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (108, '553069077', 'ANGGYA ESA MAULIDA', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (109, '552675301', 'DEA CITRA APRILIA', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (110, '553089493', 'DENIS MUHAMAD RISQI', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (111, '549171909', 'MAUDI ANUGRAH HEKSA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (112, '549191557', 'MELLA ROSE', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (113, '548948613', 'NURI MAULIDA', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (114, '550443845', 'SASKIA AQILA KAADZIYAH', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (115, '551111301', 'SHIFFA APRILIA', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (116, '551580789', 'SITI ASTRI APSIAH', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (117, '551226949', 'SITI DELA DAVINA RAMADHANI', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (118, '551168133', 'TASSYA NURMAENI', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (119, '551371365', 'TAUFIK RAU''UF ERYADI', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (120, '548084453', 'UMAR SUBAGYA SUBARKAH', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (121, '3772860212', 'WINDI MEIDI', 'Siswa', 'Hadir', 'Manual', '2026-01-07 07:00:00', '2026-01-07');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (122, '4267792717', 'ADNAN MUHAMMAD ABDULLAH', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (123, '2970990624', 'AMANDA DINAR', 'Siswa', 'Sakit', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (124, '2971077488', 'ELVIRA AZZAHRA', 'Siswa', 'Sakit', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (125, '548273589', 'FEBIYANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (126, '4268865693', 'LAILA RAMADANI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (127, '4267571757', 'LINA ANGGRAENI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (128, '549187941', 'M. HILMAN MAULANA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (129, '232410375', 'MARSYA DWI PERMANA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (130, '232410376', 'MEYSIE AYU ANGGRIANI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (131, '4268764973', 'MUHAMAD RAKHA SETIAWAN', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (132, '4267936877', 'MUHAMMAD RIJAL', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (133, '4267839821', 'MUHAMMAD RIZKY PENWAR', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (134, '4267630909', 'MUHAMMAD SYAHRUL PRATAMA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (135, '4268516381', 'NUR FAIDAH', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (136, '4267665597', 'PUTRI MELATI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (137, '2973541296', 'RISKA CAHYA ANJANI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (138, '2956973696', 'SARAH AMELIA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (139, '2970698832', 'SEPTI RAMADHANI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (140, '2970661552', 'SERLI CAHYA NURDINI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (141, '232410388', 'SHIRREN APRILIA WAHYUDI', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (142, '2970906832', 'SITI FATIMAH AZ-ZAHRA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (143, '2956391200', 'SITI HABIBAH', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (144, '2970746720', 'SITI LUTFIAH NUR SAWILYAH', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (145, '2971273952', 'SITI NUR THALITA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (146, '2971266224', 'SITI SALMA AZKIA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (147, '2970984864', 'SYALWA AS SAEFU ROMADONA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (148, '2971134512', 'TEDDY RAHMAT SUPRIATNA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (149, '2971112192', 'VEGA AULIA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (150, '4267729837', 'ZAHRA KAMILA WIJAYA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (151, '4267892717', 'ZAHRA ZASKI AULIA', 'Siswa', 'Hadir', 'Manual', '2026-02-05 07:00:00', '2026-02-05');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (152, '2970990624', 'AMANDA DINAR', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (153, '2971077488', 'ELVIRA AZZAHRA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (154, '548273589', 'FEBIYANTI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (155, '4268865693', 'LAILA RAMADANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (156, '4267571757', 'LINA ANGGRAENI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (157, '549187941', 'M. HILMAN MAULANA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (158, '232410375', 'MARSYA DWI PERMANA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (159, '232410376', 'MEYSIE AYU ANGGRIANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (160, '4268764973', 'MUHAMAD RAKHA SETIAWAN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (161, '4267936877', 'MUHAMMAD RIJAL', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (162, '4267839821', 'MUHAMMAD RIZKY PENWAR', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (163, '4267630909', 'MUHAMMAD SYAHRUL PRATAMA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (164, '4268516381', 'NUR FAIDAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (165, '4267665597', 'PUTRI MELATI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (166, '2973541296', 'RISKA CAHYA ANJANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (167, '2956973696', 'SARAH AMELIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (168, '2970698832', 'SEPTI RAMADHANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (169, '2970661552', 'SERLI CAHYA NURDINI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (170, '232410388', 'SHIRREN APRILIA WAHYUDI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (171, '2970906832', 'SITI FATIMAH AZ-ZAHRA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (172, '2956391200', 'SITI HABIBAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (173, '2970746720', 'SITI LUTFIAH NUR SAWILYAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (174, '2971273952', 'SITI NUR THALITA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (175, '2971266224', 'SITI SALMA AZKIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (176, '2970984864', 'SYALWA AS SAEFU ROMADONA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (177, '2971134512', 'TEDDY RAHMAT SUPRIATNA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (178, '2971112192', 'VEGA AULIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (179, '4267729837', 'ZAHRA KAMILA WIJAYA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (180, '4267892717', 'ZAHRA ZASKI AULIA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (181, '4267796637', 'AFRIL NUR ASSYIFA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (182, '2970675168', 'AGISTY SULISTIYU', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (183, '4267701533', 'AJENG AYUNI PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (184, '2970977792', 'ALVIANSYAH PRIASETYA NAZAR', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (185, '2970861920', 'ANANDA ASRI RAMADANTI', 'Siswa', 'Sakit', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (186, '2970371472', 'ANNISA NUR FEBRIYANTY', 'Siswa', 'Izin', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (187, '2970365152', 'AZKIA NURUL FAJARI', 'Siswa', 'Sakit', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (188, '2970994384', 'CLARA ANNA AGUSTINA', 'Siswa', 'Izin', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (189, '4268612013', 'DEASI CAHYANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (190, '2970284752', 'DESWA ADHANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (191, '2970925024', 'FAKHRI DWIANDRA MAULANDI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (192, '549167989', 'HESTINAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (193, '2970943216', 'INDAH MAYANG SARI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (194, '4267911501', 'MOH. VICKY FIRANSYA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (195, '4267839837', 'MUHAMAD ABIDIN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (196, '4267662685', 'MUHAMAD IHSAN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (197, '4267718557', 'MUHAMAD NUR WAHYU', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (198, '552846421', 'MUTIA SABILA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (199, '4267716509', 'NATASYA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (200, '4268593949', 'NAYSYILA HERAWATI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (201, '4267749965', 'PUTRI FAIHA AL - ROJAK', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (202, '2971126944', 'SITI NUR FADILLA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (203, '4268575341', 'TRISA APRILIA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (204, '2970652432', 'VISKA CANTIKA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (205, '2970888032', 'YUDISTI ZETIRA RAMADAN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (206, '4267849741', 'ZAHRA ABDULLAH', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (207, '2959012784', 'ZAQI MUHAMAD AKBAR', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (208, '4268751693', 'ZIDAN MAULIDAN', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (209, '549360453', 'AZAHRA ADELIA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (210, '551072933', 'DELLIA OCTHAVIANI', 'Siswa', 'Alpha', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (211, '552602565', 'DESWITA PUTRI ANGGRAENI', 'Siswa', 'Alpha', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (212, '552640965', 'FITRI SETIAWAN', 'Siswa', 'Alpha', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (213, '553042661', 'INGGRIT NURGRITA SARI', 'Siswa', 'Izin', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (214, '552868901', 'INKA SINTYA MAYLAIKA', 'Siswa', 'Izin', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (215, '548986965', 'KEVIN ARYANSYAH', 'Siswa', 'Izin', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (216, '549192053', 'MAYA DEFRIANTI', 'Siswa', 'Izin', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (217, '548702405', 'MUTIARA HANDAYANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (218, '551578325', 'NOURI NASHIRA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (219, '548539717', 'SALSABILA', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (220, '548678293', 'SANDYRA CENDANA JATI. P', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (221, '549291093', 'SEPTIAN RISKI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (222, '551176181', 'SHILVA RAMADHANI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (223, '551251733', 'SITI JULIANISA PUTRI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (224, '550747285', 'SITI MELATI', 'Siswa', 'Hadir', 'Manual', '2026-02-08 07:00:00', '2026-02-08');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (225, '2515254803', 'ABDULAH FATHUR RAHMAN ASSYAHIDIN', 'Siswa', 'Alpha', 'Manual', '2026-02-09 07:00:00', '2026-02-09');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (226, '2507632947', 'ABIMAYU AURORA MARANTIKA', 'Siswa', 'Sakit', 'Manual', '2026-02-09 07:00:00', '2026-02-09');
INSERT INTO attendance_logs (id, rfid_uid, user_name, user_type, status, method, timestamp, date) VALUES (227, '222310301', 'ADINDA SARAH GUMATI', 'Siswa', 'Sakit', 'Manual', '2026-02-09 07:00:00', '2026-02-09');

-- whatsapp_logs (16 rows)
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (1, '6281310783701', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 Apepsiii

Telah melakukan presensi pada: 
🕒 Pukul: 12:00 
✅ Status: Datang

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'success', '{"code":200,"messages":[{"id":"2b3a8adf-554f-4eda-8859-0bf50c79dd92","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"cc632bc2"}]}', '2026-02-05 05:00:26');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (2, '6281310783702', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 Asa

Telah melakukan presensi pada: 
🕒 Pukul: 12:00 
✅ Status: Datang

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'success', '{"code":200,"messages":[{"id":"ed90bb81-25f6-4e36-b98f-77f42bfb5f27","type":"image","to":"6281310783702@s.whatsapp.net","recipient_type":"individual","tag":"bb9f14c3"}]}', '2026-02-05 05:00:36');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (3, '6281310783701', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 Testing ah

Telah melakukan presensi pada: 
🕒 Pukul: 12:00 
✅ Status: Sakit

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'success', '{"code":200,"messages":[{"id":"b6465988-0820-4f74-b458-6b6410eb7f7a","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"246c98f2"}]}', '2026-02-05 05:00:44');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (4, '6281310783701', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 Testing ah asa

Telah melakukan presensi pada: 
🕒 Pukul: 15:25 
✅ Status: Alpha

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'success', '{"code":200,"messages":[{"id":"3f0f901c-63a0-4532-a969-5a6296c89d20","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"209f979c"}]}', '2026-02-05 08:25:19');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (5, '6281310783701', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 Apepsiii

Telah melakukan presensi pada: 
🕒 Pukul: 11:48 
✅ Status: Datang

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'success', '{"code":200,"messages":[{"id":"49051e52-da4e-4305-87e0-f434276f39a7","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"b2542a8e"}]}', '2026-02-06 04:48:48');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (6, '6281310783701', '⚠️ INFO KETERLAMBATAN SISWA

Yth. Bapak/Ibu Orang Tua/Wali,

Kami informasikan bahwa putra/putri Anda: 
👤 Nama: Testing ah 
📅 Tanggal: {date} 
⏰ Tiba Pukul: 15:17 WIB

Status kehadiran hari ini tercatat: *TERLAMBAT.*

Siswa telah diizinkan masuk kelas untuk mengikuti pelajaran. Mohon kerjasama Bapak/Ibu untuk mendorong kedatangan tepat waktu di kemudian hari.

Terima kasih. 
_— Sistem Presensi Sekolah —_', 'success', '{"code":200,"messages":[{"id":"af46a38d-74ce-4a9b-9b8f-839f6fe29336","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"4df3c77f"}]}', '2026-02-06 08:17:01');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (7, '6285158250766', '⚠️ INFO KETERLAMBATAN SISWA

Yth. Bapak/Ibu Orang Tua/Wali,

Kami informasikan bahwa putra/putri Anda: 
👤 Nama: Asa 
📅 Tanggal: {date} 
⏰ Tiba Pukul: 15:17 WIB

Status kehadiran hari ini tercatat: *TERLAMBAT.*

Siswa telah diizinkan masuk kelas untuk mengikuti pelajaran. Mohon kerjasama Bapak/Ibu untuk mendorong kedatangan tepat waktu di kemudian hari.

Terima kasih. 
_— Sistem Presensi Sekolah —_', 'success', '{"code":200,"messages":[{"id":"93e8fbbc-0ced-4a3e-bb4e-97ce89a53dac","type":"image","to":"6285158250766@s.whatsapp.net","recipient_type":"individual","tag":"eba671ef"}]}', '2026-02-06 08:17:20');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (8, '6281310783701', '⚠️ INFO KETERLAMBATAN SISWA

Yth. Bapak/Ibu Orang Tua/Wali,

Kami informasikan bahwa putra/putri Anda: 
👤 Nama: Testing ah asa 
📅 Tanggal: {date} 
⏰ Tiba Pukul: 15:17 WIB

Status kehadiran hari ini tercatat: *TERLAMBAT.*

Siswa telah diizinkan masuk kelas untuk mengikuti pelajaran. Mohon kerjasama Bapak/Ibu untuk mendorong kedatangan tepat waktu di kemudian hari.

Terima kasih. 
_— Sistem Presensi Sekolah —_', 'failed', 'Post "https://sender2.pionireradigital.id/api/v1/messages": context deadline exceeded (Client.Timeout exceeded while awaiting headers)', '2026-02-06 08:17:33');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (9, '6281310783701', '🏠 INFO KEPULANGAN SISWA

Yth. Orang Tua/Wali Testing ah,

Kami informasikan bahwa kegiatan belajar mengajar hari ini, {date}, telah selesai.

Putra/putri Anda telah melakukan presensi pulang pada pukul 15:39 WIB dan telah meninggalkan area sekolah.

Mohon dipantau kedatangannya di rumah. Terima kasih.

— Sistem Presensi Sekolah —', 'success', '{"code":200,"messages":[{"id":"b30d5847-b909-49f6-8c69-b8f8a63ee863","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"99eb68a2"}]}', '2026-02-06 08:39:34');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (10, '6285158250766', '🏠 INFO KEPULANGAN SISWA

Yth. Orang Tua/Wali Asa,

Kami informasikan bahwa kegiatan belajar mengajar hari ini, {date}, telah selesai.

Putra/putri Anda telah melakukan presensi pulang pada pukul 15:39 WIB dan telah meninggalkan area sekolah.

Mohon dipantau kedatangannya di rumah. Terima kasih.

— Sistem Presensi Sekolah —', 'success', '{"code":200,"messages":[{"id":"123595d6-affd-4d49-a735-da960374a640","type":"image","to":"6285158250766@s.whatsapp.net","recipient_type":"individual","tag":"deff2217"}]}', '2026-02-06 08:39:38');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (11, '6281310783701', '🏠 INFO KEPULANGAN SISWA

Yth. Orang Tua/Wali Testing ah asa,

Kami informasikan bahwa kegiatan belajar mengajar hari ini, 06-02-2026, telah selesai.

Putra/putri Anda telah melakukan presensi pulang pada pukul 15:48 WIB dan telah meninggalkan area sekolah.

Mohon dipantau kedatangannya di rumah. Terima kasih.

— Sistem Presensi Sekolah —', 'success', '{"code":200,"messages":[{"id":"2a9a1a2f-9298-4d20-ae10-05807fc29314","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"b5d7f998"}]}', '2026-02-06 08:48:26');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (12, '6285811158295', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 ABDULAH FATHUR RAHMAN ASSYAHIDIN

Telah melakukan presensi pada: 
🕒 Pukul: 19:06 
✅ Status: Datang

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'success', '{"code":200,"messages":[{"id":"08b7b645-a4f5-4f56-b238-966f0a773bbc","type":"image","to":"6285811158295@s.whatsapp.net","recipient_type":"individual","tag":"611a72da"}]}', '2026-02-08 12:06:13');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (13, '6281310783701', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 ABIMAYU AURORA MARANTIKA

Telah melakukan presensi pada: 
🕒 Pukul: 19:07 
✅ Status: Datang

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'success', '{"code":200,"messages":[{"id":"13fed03c-6732-4ff1-bf83-8283483718b6","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"d990970c"}]}', '2026-02-08 12:07:05');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (14, '6281310783701', '🏠 INFO KEPULANGAN SISWA

Yth. Orang Tua/Wali ABIMAYU AURORA MARANTIKA,

Kami informasikan bahwa kegiatan belajar mengajar hari ini, 08-02-2026, telah selesai.

Putra/putri Anda telah melakukan presensi pulang pada pukul 20:15 WIB dan telah meninggalkan area sekolah.

Mohon dipantau kedatangannya di rumah. Terima kasih.

— Sistem Presensi Sekolah —', 'success', '{"code":200,"messages":[{"id":"ec211158-e4a8-4ac5-a05f-eab8f5b3359a","type":"image","to":"6281310783701@s.whatsapp.net","recipient_type":"individual","tag":"af5ca1ec"}]}', '2026-02-08 13:15:41');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (15, '6281310783701', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 ADINDA SARAH GUMATI

Telah melakukan presensi pada: 
🕒 Pukul: 21:42 
✅ Status: Datang

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'failed', 'Method Not Allowed', '2026-02-08 14:42:10');
INSERT INTO whatsapp_logs (id, target, message, status, response, timestamp) VALUES (16, '6285890198663', '🔔 INFO PRESENSI SEKOLAH

Yth. Orang Tua/Wali,

Kami menginformasikan bahwa siswa atas nama: 👤 ADNAN MUHAMMAD ABDULLAH

Telah melakukan presensi pada: 
🕒 Pukul: 21:45 
✅ Status: Datang

Terima kasih atas perhatiannya.

Pesan ini dikirim otomatis oleh sistem.', 'failed', 'Method Not Allowed', '2026-02-08 14:45:17');

-- prayer_logs (72 rows)
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (1, '1', 'Testing ah', 'X - MPLB', 'Ashar', '2026-02-06 15:17:06', '2026-02-06', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (2, '2', 'Asa', 'X - MPLB', 'Ashar', '2026-02-06 15:17:11', '2026-02-06', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (3, '3', 'Apepsiii', 'X - MPLB', 'Ashar', '2026-02-06 15:17:14', '2026-02-06', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (4, '4', 'Testing ah asa', 'X - MPLB', 'Ashar', '2026-02-06 15:17:16', '2026-02-06', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (5, '3730162260', 'ADJMA NURPIRDA', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (6, '553069077', 'ANGGYA ESA MAULIDA', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (7, '550443845', 'SASKIA AQILA KAADZIYAH', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (8, '551111301', 'SHIFFA APRILIA', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (9, '551580789', 'SITI ASTRI APSIAH', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (10, '548084453', 'UMAR SUBAGYA SUBARKAH', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (11, '551371365', 'TAUFIK RAU''UF ERYADI', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (12, '550791109', 'ALISA APRIL YANTI', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (13, '553089493', 'DENIS MUHAMAD RISQI', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (14, '548948613', 'NURI MAULIDA', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (15, '551226949', 'SITI DELA DAVINA RAMADHANI', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (16, '551168133', 'TASSYA NURMAENI', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (17, '3772860212', 'WINDI MEIDI', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (18, '3729794356', 'AGIESKHA VIANA SIAMY', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (19, '552675301', 'DEA CITRA APRILIA', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (20, '549171909', 'MAUDI ANUGRAH HEKSA PUTRI', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (21, '549191557', 'MELLA ROSE', 'XI-MPLB', 'Dzuhur', '2026-02-09 23:52:44', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (22, '3730162260', 'ADJMA NURPIRDA', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (23, '3729794356', 'AGIESKHA VIANA SIAMY', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (24, '550791109', 'ALISA APRIL YANTI', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (25, '553069077', 'ANGGYA ESA MAULIDA', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (26, '551371365', 'TAUFIK RAU''UF ERYADI', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (27, '548084453', 'UMAR SUBAGYA SUBARKAH', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (28, '552675301', 'DEA CITRA APRILIA', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (29, '553089493', 'DENIS MUHAMAD RISQI', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (30, '549191557', 'MELLA ROSE', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (31, '548948613', 'NURI MAULIDA', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (32, '551111301', 'SHIFFA APRILIA', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (33, '3772860212', 'WINDI MEIDI', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (34, '549171909', 'MAUDI ANUGRAH HEKSA PUTRI', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (35, '550443845', 'SASKIA AQILA KAADZIYAH', 'XI-MPLB', 'Ashar', '2026-02-09 23:53:00', '2026-02-09', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (36, '549171909', 'MAUDI ANUGRAH HEKSA PUTRI', 'XI-MPLB', 'Ashar', '2026-02-10 23:53:24', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (37, '3730162260', 'ADJMA NURPIRDA', 'XI-MPLB', 'Ashar', '2026-02-10 23:53:24', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (38, '3729794356', 'AGIESKHA VIANA SIAMY', 'XI-MPLB', 'Ashar', '2026-02-10 23:53:24', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (39, '550791109', 'ALISA APRIL YANTI', 'XI-MPLB', 'Ashar', '2026-02-10 23:53:24', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (40, '552675301', 'DEA CITRA APRILIA', 'XI-MPLB', 'Ashar', '2026-02-10 23:53:24', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (41, '553089493', 'DENIS MUHAMAD RISQI', 'XI-MPLB', 'Ashar', '2026-02-10 23:53:24', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (42, '3729794356', 'AGIESKHA VIANA SIAMY', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (43, '548084453', 'UMAR SUBAGYA SUBARKAH', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (44, '551226949', 'SITI DELA DAVINA RAMADHANI', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (45, '3730162260', 'ADJMA NURPIRDA', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (46, '548948613', 'NURI MAULIDA', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (47, '550443845', 'SASKIA AQILA KAADZIYAH', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (48, '551111301', 'SHIFFA APRILIA', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (49, '551580789', 'SITI ASTRI APSIAH', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (50, '553069077', 'ANGGYA ESA MAULIDA', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (51, '549191557', 'MELLA ROSE', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (52, '551371365', 'TAUFIK RAU''UF ERYADI', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (53, '3772860212', 'WINDI MEIDI', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (54, '551168133', 'TASSYA NURMAENI', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (55, '550791109', 'ALISA APRIL YANTI', 'XI-MPLB', 'Dzuhur', '2026-02-10 23:53:35', '2026-02-10', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (56, '3730162260', 'ADJMA NURPIRDA', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (57, '3729794356', 'AGIESKHA VIANA SIAMY', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (58, '552675301', 'DEA CITRA APRILIA', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (59, '549191557', 'MELLA ROSE', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'PMS', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (60, '550443845', 'SASKIA AQILA KAADZIYAH', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (61, '551226949', 'SITI DELA DAVINA RAMADHANI', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (62, '551168133', 'TASSYA NURMAENI', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (63, '550791109', 'ALISA APRIL YANTI', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (64, '553069077', 'ANGGYA ESA MAULIDA', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'PMS', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (65, '548084453', 'UMAR SUBAGYA SUBARKAH', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (66, '553089493', 'DENIS MUHAMAD RISQI', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (67, '551111301', 'SHIFFA APRILIA', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (68, '551371365', 'TAUFIK RAU''UF ERYADI', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (69, '549171909', 'MAUDI ANUGRAH HEKSA PUTRI', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (70, '548948613', 'NURI MAULIDA', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (71, '551580789', 'SITI ASTRI APSIAH', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'Hadir', 'RFID');
INSERT INTO prayer_logs (id, rfid_uid, name, class_name, prayer_type, timestamp, date, status, recorded_by) VALUES (72, '3772860212', 'WINDI MEIDI', 'XI-MPLB', 'Dzuhur', '2026-02-11 00:10:37', '2026-02-11', 'PMS', 'RFID');