SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

CREATE TABLE IF NOT EXISTS `motify_agents` (
`id_agent` int(11) NOT NULL,
  `a_fk_integration` int(11) NOT NULL,
  `a_name` varchar(100) NOT NULL,
  `a_company_id` varchar(50) NOT NULL,
  `a_description` varchar(255) NOT NULL,
  `a_logo` varchar(255) NOT NULL,
  `a_bg_image` varchar(255) NOT NULL,
  `a_address` varchar(255) NOT NULL,
  `a_phone` varchar(50) NOT NULL,
  `a_email` varchar(255) NOT NULL,
  `a_site` varchar(255) NOT NULL,
  `a_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `a_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DELIMITER //
CREATE TRIGGER `motify_agents_a_updated_at` BEFORE UPDATE ON `motify_agents`
 FOR EACH ROW BEGIN SET NEW.A_UPDATED_AT = CURRENT_TIMESTAMP; END
//
DELIMITER ;

CREATE TABLE IF NOT EXISTS `motify_agent_employees` (
`id_employee` int(11) NOT NULL,
  `e_fk_agent` int(11) NOT NULL,
  `e_fk_user` int(11) DEFAULT NULL,
  `e_code` varchar(50) NOT NULL,
  `e_name` varchar(255) NOT NULL,
  `e_role` varchar(255) NOT NULL,
  `e_email` varchar(255) NOT NULL,
  `e_hire_date` date NOT NULL,
  `e_number_of_dependants` int(2) NOT NULL DEFAULT '0',
  `e_gross_base_salary` float NOT NULL,
  `e_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `e_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DELIMITER //
CREATE TRIGGER `motify_agent_employees_e_updated_at` BEFORE UPDATE ON `motify_agent_employees`
 FOR EACH ROW BEGIN
	SET NEW.E_UPDATED_AT = CURRENT_TIMESTAMP;
END
//
DELIMITER ;

CREATE TABLE IF NOT EXISTS `motify_agent_settings` (
`id_setting` int(11) NOT NULL,
  `s_fk_agent` int(11) NOT NULL,
  `s_fk_user` int(11) DEFAULT NULL,
  `s_fk_agent_processed` int(11) DEFAULT NULL,
  `s_role` varchar(255) NOT NULL,
  `s_notifications_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `s_is_main_agent` tinyint(1) NOT NULL DEFAULT '0',
  `s_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `s_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DELIMITER //
CREATE TRIGGER `motify_agent_settings_s_updated_at` BEFORE UPDATE ON `motify_agent_settings`
 FOR EACH ROW BEGIN SET NEW.S_UPDATED_AT = CURRENT_TIMESTAMP; END
//
DELIMITER ;

CREATE TABLE IF NOT EXISTS `motify_integrations` (
`id_integration` int(11) NOT NULL,
  `i_hash` varchar(32) NOT NULL,
  `i_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `i_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DELIMITER //
CREATE TRIGGER `motify_integrations_i_updated_at` BEFORE UPDATE ON `motify_integrations`
 FOR EACH ROW BEGIN SET NEW.I_UPDATED_AT = CURRENT_TIMESTAMP; END
//
DELIMITER ;

CREATE TABLE IF NOT EXISTS `motify_payslip` (
`id_payslip` int(11) NOT NULL,
  `p_fk_employee` int(11) NOT NULL,
  `p_title` varchar(100) NOT NULL,
  `p_currency` varchar(10) NOT NULL,
  `p_amount` float NOT NULL,
  `p_data` text NOT NULL,
  `p_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `p_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DELIMITER //
CREATE TRIGGER `motify_payslip_p_updated_at` BEFORE UPDATE ON `motify_payslip`
 FOR EACH ROW BEGIN SET NEW.P_UPDATED_AT = CURRENT_TIMESTAMP; END
//
DELIMITER ;

CREATE TABLE IF NOT EXISTS `motify_users` (
`id_user` int(11) NOT NULL,
  `u_fk_integration` int(11) DEFAULT NULL,
  `u_name` varchar(255) NOT NULL,
  `u_description` varchar(255) NOT NULL,
  `u_short` varchar(255) NOT NULL,
  `u_avatar` varchar(255) NOT NULL,
  `u_phone` varchar(50) NOT NULL,
  `u_email` varchar(255) NOT NULL,
  `u_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `u_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DELIMITER //
CREATE TRIGGER `motify_users_u_updated_at` BEFORE UPDATE ON `motify_users`
 FOR EACH ROW BEGIN SET NEW.U_UPDATED_AT = CURRENT_TIMESTAMP; END
//
DELIMITER ;

CREATE TABLE IF NOT EXISTS `motify_user_access` (
`id_user_access` int(11) NOT NULL,
  `ua_fk_user` int(11) NOT NULL,
  `ua_type` enum('email','fb','vk') NOT NULL,
  `ua_email` varchar(255) DEFAULT NULL,
  `ua_phone` varchar(50) DEFAULT NULL,
  `ua_password` varchar(255) NOT NULL,
  `ua_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `ua_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
DELIMITER //
CREATE TRIGGER `motify_user_access_ua_updated_at` BEFORE UPDATE ON `motify_user_access`
 FOR EACH ROW BEGIN SET NEW.UA_UPDATED_AT = CURRENT_TIMESTAMP; END
//
DELIMITER ;


ALTER TABLE `motify_agents`
 ADD PRIMARY KEY (`id_agent`), ADD KEY `a_fk_integration` (`a_fk_integration`);

ALTER TABLE `motify_agent_employees`
 ADD PRIMARY KEY (`id_employee`), ADD UNIQUE KEY `e_uniq_fk_agent_fk_user` (`e_fk_agent`,`e_fk_user`), ADD KEY `e_fk_user` (`e_fk_user`), ADD KEY `e_fk_agent` (`e_fk_agent`);

ALTER TABLE `motify_agent_settings`
 ADD PRIMARY KEY (`id_setting`), ADD UNIQUE KEY `s_uniq_fk_agent_fk_user` (`s_fk_agent`,`s_fk_user`), ADD KEY `s_fk_user` (`s_fk_user`), ADD KEY `s_fk_agent` (`s_fk_agent`), ADD KEY `s_fk_agent_processed` (`s_fk_agent_processed`);

ALTER TABLE `motify_integrations`
 ADD PRIMARY KEY (`id_integration`), ADD UNIQUE KEY `i_hash` (`i_hash`);

ALTER TABLE `motify_payslip`
 ADD PRIMARY KEY (`id_payslip`), ADD KEY `p_fk_employee` (`p_fk_employee`);

ALTER TABLE `motify_users`
 ADD PRIMARY KEY (`id_user`), ADD KEY `u_fk_integration` (`u_fk_integration`);

ALTER TABLE `motify_user_access`
 ADD PRIMARY KEY (`id_user_access`), ADD UNIQUE KEY `ua_uniq_type_fk_user` (`ua_type`,`ua_fk_user`), ADD KEY `ua_fk_user` (`ua_fk_user`);


ALTER TABLE `motify_agents`
MODIFY `id_agent` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `motify_agent_employees`
MODIFY `id_employee` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `motify_agent_settings`
MODIFY `id_setting` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `motify_integrations`
MODIFY `id_integration` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `motify_payslip`
MODIFY `id_payslip` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `motify_users`
MODIFY `id_user` int(11) NOT NULL AUTO_INCREMENT;
ALTER TABLE `motify_user_access`
MODIFY `id_user_access` int(11) NOT NULL AUTO_INCREMENT;

ALTER TABLE `motify_agents`
ADD CONSTRAINT `motify_agents_ibfk_1` FOREIGN KEY (`a_fk_integration`) REFERENCES `motify_integrations` (`id_integration`) ON UPDATE CASCADE;

ALTER TABLE `motify_agent_employees`
ADD CONSTRAINT `e_motify_agent_employees_ibfk_1` FOREIGN KEY (`e_fk_agent`) REFERENCES `motify_agents` (`id_agent`) ON UPDATE CASCADE,
ADD CONSTRAINT `e_motify_agent_employees_ibfk_2` FOREIGN KEY (`e_fk_user`) REFERENCES `motify_users` (`id_user`) ON UPDATE CASCADE;

ALTER TABLE `motify_agent_settings`
ADD CONSTRAINT `s_motify_agent_settings_ibfk_1` FOREIGN KEY (`s_fk_agent`) REFERENCES `motify_agents` (`id_agent`) ON UPDATE CASCADE,
ADD CONSTRAINT `s_motify_agent_settings_ibfk_2` FOREIGN KEY (`s_fk_user`) REFERENCES `motify_users` (`id_user`) ON UPDATE CASCADE,
ADD CONSTRAINT `s_motify_agent_settings_ibfk_3` FOREIGN KEY (`s_fk_agent_processed`) REFERENCES `motify_agents` (`id_agent`) ON UPDATE CASCADE;

ALTER TABLE `motify_payslip`
ADD CONSTRAINT `p_motify_payslip_ibfk_1` FOREIGN KEY (`p_fk_employee`) REFERENCES `motify_agent_employees` (`id_employee`) ON UPDATE CASCADE;

ALTER TABLE `motify_users`
ADD CONSTRAINT `motify_users_ibfk_1` FOREIGN KEY (`u_fk_integration`) REFERENCES `motify_integrations` (`id_integration`) ON UPDATE CASCADE;

ALTER TABLE `motify_user_access`
ADD CONSTRAINT `ua_motify_user_access_ibfk_1` FOREIGN KEY (`ua_fk_user`) REFERENCES `motify_users` (`id_user`) ON UPDATE CASCADE;

-- migration1

ALTER TABLE `motify_users` ADD `u_email_approved` BOOLEAN NOT NULL AFTER `u_email`, ADD `u_phone_approved` BOOLEAN NOT NULL AFTER `u_email_approved`;

-- migration2

ALTER TABLE `motify_user_access` CHANGE `ua_email` `ua_login` VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL;
ALTER TABLE `motify_user_access` DROP `ua_phone`;
ALTER TABLE `motify_user_access` CHANGE `ua_type` `ua_type` ENUM('email','fb','google','phone') CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL;

-- migration3

DROP TABLE IF EXISTS `motify_device`;
CREATE TABLE `motify_device` (
  `id_device` int(11) NOT NULL,
  `d_fk_user` int(11) NOT NULL,
  `d_token` varchar(255) NOT NULL,
  `d_device` varchar(255) NOT NULL,
  `d_updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `d_created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


ALTER TABLE `motify_device`
  ADD PRIMARY KEY (`id_device`),
  ADD KEY `d_fk_user_ind` (`d_fk_user`),
  ADD UNIQUE( `d_token`);


ALTER TABLE `motify_device`
  MODIFY `id_device` int(11) NOT NULL AUTO_INCREMENT;


ALTER TABLE `motify_device`
  ADD CONSTRAINT `d_motify_device_ibfk_1` FOREIGN KEY (`d_fk_user`) REFERENCES `motify_users` (`id_user`) ON DELETE CASCADE ON UPDATE CASCADE;
COMMIT;

