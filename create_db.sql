create database motify_core_api;

-- phpMyAdmin SQL Dump
-- version 4.2.8.1
-- http://www.phpmyadmin.net
--
-- Хост: localhost
-- Время создания: Мар 10 2018 г., 00:23
-- Версия сервера: 5.6.19
-- Версия PHP: 5.4.45

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- База данных: `motify_core_api`
--
CREATE DATABASE IF NOT EXISTS `motify_core_api` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `motify_core_api`;

-- --------------------------------------------------------

--
-- Структура таблицы `motify_agents`
--

CREATE TABLE IF NOT EXISTS `motify_agents` (
`id_agent` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `company_id` varchar(50) NOT NULL,
  `desc` varchar(255) NOT NULL,
  `address` varchar(255) NOT NULL,
  `phone` varchar(50) NOT NULL,
  `email` varchar(255) NOT NULL,
  `site` varchar(255) NOT NULL,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `motify_agent_employees`
--

CREATE TABLE IF NOT EXISTS `motify_agent_employees` (
  `fk_agent` int(11) NOT NULL,
  `fk_user` int(11) NOT NULL,
  `employee_code` varchar(50) NOT NULL,
  `hire_date` date DEFAULT NULL,
  `number_of_dependants` int(2) NOT NULL DEFAULT '0',
  `gross_base_salary` float NOT NULL,
  `role` varchar(255) NOT NULL,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `motify_agent_settings`
--

CREATE TABLE IF NOT EXISTS `motify_agent_settings` (
  `fk_agent` int(11) NOT NULL,
  `fk_user` int(11) NOT NULL,
  `role` varchar(255) NOT NULL,
  `notifications_enabled` tinyint(1) NOT NULL DEFAULT '0',
  `is_main_agent` tinyint(1) NOT NULL DEFAULT '0',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `motify_payslip`
--

CREATE TABLE IF NOT EXISTS `motify_payslip` (
`id_payslip` int(11) NOT NULL,
  `fk_employee` int(11) NOT NULL,
  `currency` varchar(10) NOT NULL,
  `amount` float NOT NULL,
  `data` text NOT NULL,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `motify_users`
--

CREATE TABLE IF NOT EXISTS `motify_users` (
`id_user` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `p_description` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `phone` varchar(50) NOT NULL,
  `email` varchar(255) NOT NULL,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Структура таблицы `motify_user_access`
--

CREATE TABLE IF NOT EXISTS `motify_user_access` (
`id_user_access` int(11) NOT NULL,
  `type_access` enum('email','fb','vk') NOT NULL,
  `email` varchar(255) NOT NULL,
  `phone` varchar(50) NOT NULL,
  `fk_user` int(11) NOT NULL,
  `password` varchar(255) NOT NULL,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Индексы сохранённых таблиц
--

--
-- Индексы таблицы `motify_agents`
--
ALTER TABLE `motify_agents`
 ADD PRIMARY KEY (`id_agent`);

--
-- Индексы таблицы `motify_agent_employees`
--
ALTER TABLE `motify_agent_employees`
 ADD PRIMARY KEY (`fk_agent`,`fk_user`), ADD UNIQUE KEY `id_agent` (`fk_agent`,`fk_user`);

--
-- Индексы таблицы `motify_agent_settings`
--
ALTER TABLE `motify_agent_settings`
 ADD PRIMARY KEY (`fk_agent`,`fk_user`), ADD UNIQUE KEY `id_agent` (`fk_agent`,`fk_user`);

--
-- Индексы таблицы `motify_payslip`
--
ALTER TABLE `motify_payslip`
 ADD PRIMARY KEY (`id_payslip`);

--
-- Индексы таблицы `motify_users`
--
ALTER TABLE `motify_users`
 ADD PRIMARY KEY (`id_user`);

--
-- Индексы таблицы `motify_user_access`
--
ALTER TABLE `motify_user_access`
 ADD PRIMARY KEY (`id_user_access`);

--
-- AUTO_INCREMENT для сохранённых таблиц
--

--
-- AUTO_INCREMENT для таблицы `motify_agents`
--
ALTER TABLE `motify_agents`
MODIFY `id_agent` int(11) NOT NULL AUTO_INCREMENT;
--
-- AUTO_INCREMENT для таблицы `motify_payslip`
--
ALTER TABLE `motify_payslip`
MODIFY `id_payslip` int(11) NOT NULL AUTO_INCREMENT;
--
-- AUTO_INCREMENT для таблицы `motify_users`
--
ALTER TABLE `motify_users`
MODIFY `id_user` int(11) NOT NULL AUTO_INCREMENT;
--
-- AUTO_INCREMENT для таблицы `motify_user_access`
--
ALTER TABLE `motify_user_access`
MODIFY `id_user_access` int(11) NOT NULL AUTO_INCREMENT;
