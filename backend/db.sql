/*
SQLyog Ultimate v12.5.0 (64 bit)
MySQL - 5.7.16-log : Database - guo_depot
*********************************************************************
*/

/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
CREATE DATABASE /*!32312 IF NOT EXISTS*/`guo_depot` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */;

USE `guo_depot`;

/*Table structure for table `admins` */

DROP TABLE IF EXISTS `admins`;

CREATE TABLE `admins` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` bigint(20) DEFAULT '1',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_admins_username` (`username`),
  KEY `idx_admins_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

/*Data for the table `admins` */

insert  into `admins`(`id`,`username`,`password`,`email`,`status`,`created_at`,`updated_at`,`deleted_at`) values 
(1,'admin','8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92','admin@aichat199.com',1,'2025-11-29 06:51:13.589','2025-11-29 06:51:13.589',NULL);

/*Table structure for table `cards` */

DROP TABLE IF EXISTS `cards`;

CREATE TABLE `cards` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `value` bigint(20) DEFAULT '0',
  `status` bigint(20) DEFAULT '0',
  `used_by` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `used_at` bigint(20) DEFAULT NULL,
  `expired_at` bigint(20) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_cards_code` (`code`),
  KEY `idx_cards_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

/*Data for the table `cards` */

/*Table structure for table `options` */

DROP TABLE IF EXISTS `options`;

CREATE TABLE `options` (
  `key` varchar(191) NOT NULL,
  `value` longtext,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

/*Data for the table `options` */

insert  into `options`(`key`,`value`) values 
('About','<h2 id=\"h2--\" style=\"text-wrap: wrap;\">\n    <span style=\"font-family: 微软雅黑, &quot;Microsoft YaHei&quot;;\">联系我们</span>\n</h2>\n<p style=\"text-wrap: wrap;\">\n    <span style=\"font-family: 微软雅黑, &quot;Microsoft YaHei&quot;;\">售后Q群: 511903990<br/>TG: &nbsp;</span><a href=\"https://t.me/aichat199\" style=\"font-family: 微软雅黑, &quot;Microsoft YaHei&quot;; text-decoration: underline;\"><span style=\"font-family: 微软雅黑, &quot;Microsoft YaHei&quot;;\">https://t.me/aichat199</span></a><br/><span style=\"font-family: 微软雅黑, &quot;Microsoft YaHei&quot;;\">WX：aichat199</span>\n</p>\n<p>\n    <br/>\n</p>'),
('ApproximateTokenEnabled','false'),
('AutomaticDisableChannelEnabled','true'),
('CacheChannelEnabled','false'),
('ChatLink','https://aichat199.com/'),
('CompletionRatio','{\n  \"gpt-3.5-turbo\": 3,\n  \"gpt-3.5-turbo-0613\": 5,\n  \"gpt-3.5-turbo-16k-0613\": 5,\n  \"gpt-3.5-turbo-0125\": 5,\n  \"gpt-4-turbo-2024-04-09\": 3,\n  \"gpt-4-turbo\": 3,\n  \"gpt-4o\": 4,\n  \"gpt-4o-2024-05-13\": 4,\n  \"gpt-4o-2024-08-06\": 4,\n  \"gpt-4o-2024-11-20\": 4,\n  \"codex-mini-latest\": 4,\n  \"gpt-4-1106-preview\": 3,\n  \"gpt-4-1106-vision-preview\": 3,\n  \"gpt-4-0314\": 5,\n  \"chatgpt-4o-latest\": 3,\n  \"gpt-4o-mini\": 4,\n  \"gpt-4o-mini-2024-07-18\": 4,\n  \"gpt-4o-search-preview\": 4,\n  \"gpt-4o-search-preview-2025-03-11\": 4,\n  \"gpt-4o-mini-search-preview\": 4,\n  \"gpt-4o-mini-search-preview-2025-03-11\": 4,\n  \"computer-use-preview\": 4,\n  \"computer-use-preview-2025-03-11\": 4,\n  \"gpt-4.1\": 4,\n  \"gpt-4.1-2025-04-14\": 4,\n  \"gpt-4.1-mini\": 4,\n  \"gpt-4.1-nano\": 4,\n  \"gpt-4.1-2025-04-14\": 4,\n  \"gpt-4.1-mini-2025-04-14\": 4,\n  \"gpt-4.1-nano-2025-04-14\": 4,\n  \"gpt-image-1\": 4,\n  \"gpt-4.5-preview\": 2,\n  \"gpt-4.5-preview-2025-02-27\": 2,\n  \"gpt-5\": 8,\n  \"gpt-5-2025-08-07\": 8,\n  \"gpt-5-mini\": 8,\n  \"gpt-5-mini-2025-08-07\": 8,\n  \"gpt-5-nano\": 8,\n  \"gpt-5-nano-2025-08-07\": 8,\n  \"o1\": 4,\n  \"o1-2024-12-17\": 4,\n  \"o1-preview\": 4,\n  \"o1-preview-2024-09-12\": 4,\n  \"o1-mini\": 4,\n  \"o1-mini-2024-09-12\": 4,\n  \"o1-pro\": 4,\n  \"o1-pro-2025-03-19\": 4,\n  \"o3-mini\": 4,\n  \"o3-mini-2025-01-31\": 4,\n  \"o3\": 4,\n  \"o3-2025-04-16\": 4,\n  \"o3-deep-research\": 4,\n  \"o3-deep-research-2025-06-26\": 4,\n  \"o3-pro\": 4,\n  \"o3-pro-2025-06-10\":4,\n  \"o4-mini\": 4,\n  \"o4-mini-2025-04-16\": 4,\n  \"claude-3-haiku-20240307\": 5,\n  \"claude-3-sonnet-20240229\": 5,\n  \"claude-3-opus-20240229\": 5,\n  \"claude-3-5-sonnet-20240620\": 5,\n  \"claude-3-5-sonnet-latest\": 5,\n  \"claude-3-7-sonnet-20250219\": 5,\n  \"claude-3-7-sonnet-latest\": 5,\n  \"claude-3-5-haiku-20241022\": 5,\n  \"claude-sonnet-4-20250514\": 5,\n  \"claude-opus-4-20250514\": 5,\n  \"claude-opus-4-1-20250805\": 5,\n  \"gemini-1.0-pro-vision\": 3,\n  \"gemini-1.0-pro-latest\": 4,\n  \"gemini-1.5-pro\": 4,\n  \"gemini-1.5-pro-001\": 4,\n  \"gemini-1.5-pro-002\": 4,\n  \"gemini-1.5-pro-latest\": 4,\n  \"gemini-1.5-flash\": 0.04,\n  \"gemini-1.5-flash-001\": 4,\n  \"gemini-1.5-flash-001-tuning\": 4,\n  \"gemini-1.5-flash-002\": 4,\n  \"gemini-1.5-flash-latest\": 4,\n  \"gemini-1.5-flash-exp-0827\": 4,\n  \"gemini-1.5-flash-8b\": 4,\n  \"gemini-1.5-flash-8b-001\": 4,\n  \"gemini-1.5-flash-8b-exp-0924\": 4,\n  \"gemini-1.5-flash-8b-latest\": 4,\n  \"gemini-exp-1206\": 4,\n  \"learnlm-1.5-pro-experimental\": 4,\n  \"gemini-2.0-flash\": 4,\n  \"gemini-2.0-flash-001\": 4,\n  \"gemini-2.0-flash-exp\": 4,\n  \"gemini-2.0-flash-thinking-exp\": 4,\n  \"gemini-2.0-flash-thinking-exp-01-21\": 4,\n  \"gemini-2.0-flash-lite-preview-02-05\": 4,\n  \"gemini-2.0-flash-lite-preview\": 4,\n  \"gemini-2.0-flash-exp-image-generation\": 4,\n  \"gemini-2.0-flash-preview-image-generation\": 4,\n  \"gemini-2.5-flash-image-preview\": 100,\n  \"gemini-2.0-pro-exp-02-05\": 4,\n  \"gemini-2.0-flash-lite\":  4,\n  \"gemini-2.0-flash-lite-001\": 4,\n  \"gemini-2.5-flash-lite-preview-06-17\": 8,\n  \"gemini-2.5-flash-preview-04-17\": 4,\n  \"gemini-2.5-flash-preview-04-17-thinking\": 23.33,\n  \"gemini-2.5-flash-preview-05-20\": 8,\n  \"gemini-2.5-flash-preview-05-20\": 8,\n  \"gemini-2.5-flash-preview-tts\": 20,\n  \"gemini-2.5-flash\": 8,\n  \"gemma-3-27b-it\": 4,\n  \"gemini-2.5-pro-preview-tts\": 20,\n  \"gemini-2.5-pro-exp-03-25\": 4,\n  \"gemini-2.5-pro-preview-03-25\": 8,\n  \"gemini-2.5-pro-preview-05-06\": 8,\n  \"gemini-2.5-pro-preview-06-05\": 8,\n  \"gemini-2.5-pro\": 8, \n  \"imagen-3.0-generate-002\": 4,\n  \"veo-2.0-generate-001\": 4,\n  \"deepseek-chat\": 3,\n  \"deepseek-coder\": 3,\n  \"deepseek-reasoner\": 3,\n  \"qwen-max\": 4,\n  \"qwen-max-latest\": 4,\n  \"qwen-max-longcontext\": 4,\n  \"qwen2.5-max\": 4,\n  \"qwen-plus\": 3,\n  \"qwen-plus-latest\": 3,\n  \"qwen-turbo\": 4,\n  \"qwen-turbo-latest\": 4\n}'),
('DemoSiteEnabled','false'),
('EmailVerificationEnabled','false'),
('fetch_setting.allowed_ports','[\"80\",\"443\",\"8080\",\"8443\"]'),
('fetch_setting.domain_filter_mode','false'),
('fetch_setting.domain_list','[]'),
('fetch_setting.ip_filter_mode','false'),
('fetch_setting.ip_list','[]'),
('Footer',''),
('GroupRatio','{\n  \"default\": 1,\n  \"vip\": 1,\n  \"svip\": 1\n}'),
('HomePageContent',''),
('HttpProxy','http://xiaoguo:Ji6dft4Cqd9l_eX6h3@199.119.138.75:1080'),
('Logo','https://img.alicdn.com/imgextra/i2/3566933417/O1CN01WDyh0I1b72BCVb6Co_!!3566933417.png_120x120.jpg'),
('ModelRatio','{\n  \"gpt-3.5-turbo\": 0.25,\n  \"gpt-3.5-turbo-0613\": 1.5,\n  \"gpt-3.5-turbo-1106\": 0.5,\n  \"gpt-3.5-turbo-0125\": 0.25,\n  \"gpt-3.5-turbo-16k-0613\": 1.5,\n  \"gpt-3.5-turbo-instruct\": 0.75,\n  \"gpt-4\": 15,\n  \"gpt-4-0314\": 8,\n  \"gpt-4-0613\": 15,\n  \"davinci-002\": 10,\n  \"text-embedding-ada-002\": 0.05,\n  \"text-embedding-3-small\": 0.01,\n  \"text-embedding-3-large\": 0.065,\n  \"text-moderation-latest\": 0.1,\n  \"text-moderation-stable\": 0.1,\n  \"omni-moderation\": 0.1,\n  \"codex-mini-latest\": 0.75,\n  \"dall-e-2\": 8,\n  \"dall-e-3\": 20,\n  \"gpt-4-0125-preview\": 5,\n  \"gpt-4-turbo\": 5,\n  \"gpt-4-turbo-preview\": 5,\n  \"gpt-4-1106-preview\": 8,\n  \"gpt-4-turbo-2024-04-09\": 5,\n  \"gpt-4o\": 1.25,\n  \"gpt-4o-2024-05-13\": 1.25,\n  \"gpt-4o-2024-08-06\": 1.25,\n  \"gpt-4o-2024-11-20\": 1.25,\n  \"chatgpt-4o-latest\": 1.25,\n  \"gpt-4o-mini\": 0.075,\n  \"gpt-4o-mini-2024-07-18\": 0.075,\n  \"gpt-4o-mini-search-preview\": 0.075,\n  \"gpt-4o-mini-search-preview-2025-03-11\": 0.075,\n  \"gpt-4o-mini-search-preview\": 0.075,\n  \"gpt-4o-mini-search-preview-2025-03-11\": 0.075,\n  \"gpt-4o-search-preview\": 1.25,\n  \"gpt-4o-search-preview-2025-03-11\": 1.25,\n  \"computer-use-preview\": 1.5,\n  \"computer-use-preview-2025-03-11\": 1.5,\n  \"o1-preview\": 7.5,\n  \"o1-preview-2024-09-12\": 7.5,\n  \"o1\": 8,\n  \"o1-2024-12-17\": 8,\n  \"o1-mini\": 0.55,\n  \"o1-mini-2024-09-12\": 0.55,\n  \"o1-pro\": 75,\n  \"o1-pro-2025-03-19\": 75,\n  \"o3-mini\": 0.55,\n  \"o3-mini-2025-01-31\": 0.55,\n  \"o3\": 1,\n  \"o3-2025-04-16\": 1,\n  \"o3-deep-research\": 5,\n  \"o3-deep-research-2025-06-26\": 5,\n  \"o3-pro\": 10,\n  \"o3-pro-2025-06-10\":10,\n  \"o3-2025-04-16\": 5,\n  \"o4-mini\": 0.55,\n  \"o4-mini-2025-04-16\": 0.55,\n  \"gpt-4.1\": 1,\n  \"gpt-4.1-2025-04-14\": 1,\n  \"gpt-4.1-mini\": 0.2,\n  \"gpt-4.1-mini-2025-04-14\": 0.2,\n  \"gpt-4.1-nano\": 0.05,\n  \"gpt-4.1-nano-2025-04-14\": 0.05,\n  \"gpt-4.1-2025-04-14\": 1,\n  \"gpt-4.5-preview\": 37.5,\n  \"gpt-4.5-preview-2025-02-27\": 37.5,\n  \"gpt-5\": 0.625,\n  \"gpt-5-2025-08-07\": 0.625,\n  \"gpt-5-mini\": 0.125,\n  \"gpt-5-mini-2025-08-07\": 0.125,\n  \"gpt-5-nano\": 0.025,\n  \"gpt-5-nano-2025-08-07\": 0.025,\n  \"gpt-5-chat-latest\": 0.625,\n  \"gpt-image-1\": 5,\n  \"whisper-1\": 1200,\n  \"tts-1\": 30,\n  \"tts-1-1106\": 30,\n  \"tts-1-hd\": 60,\n  \"tts-1-hd-1106\": 60,\n  \"Precise-offline\": 15,\n  \"Precise-18k-offline\": 30,\n  \"Precise-g4t-offline\": 5,\n  \"Precise-g4t-vision\": 5,\n  \"text-embedding-004\": 0.065,\n  \"gemini-embedding-exp-03-07\": 0.065,\n  \"gemini-1.0-pro-vision\": 0.25,\n  \"gemini-1.0-pro-latest\": 0.25,\n  \"gemini-1.5-pro\": 1.25,\n  \"gemini-1.5-pro-001\": 1.25,\n  \"gemini-1.5-pro-002\": 1.25,\n  \"gemini-1.5-pro-latest\": 1.25,\n  \"gemini-1.5-flash\": 0.075,\n  \"gemini-1.5-flash-001\": 0.075,\n  \"gemini-1.5-flash-001-tuning\": 0.075,\n  \"gemini-1.5-flash-002\": 0.075,\n  \"gemini-1.5-flash-latest\": 0.075,\n  \"gemini-1.5-flash-exp-0827\": 0.075,\n  \"gemini-1.5-flash-8b\": 0.0375,\n  \"gemini-1.5-flash-8b-001\": 0.0375,\n  \"gemini-1.5-flash-8b-exp-0924\": 0.0375,\n  \"gemini-1.5-flash-8b-latest\": 0.0375,\n  \"gemini-exp-1206\": 1.25,\n  \"learnlm-1.5-pro-experimental\": 1.25,\n  \"gemini-2.0-flash\": 0.35,\n  \"gemini-2.0-flash-001\": 0.35,\n  \"gemini-2.0-flash-exp\": 1.25,\n  \"gemini-2.0-flash-thinking-exp\": 1.25,\n  \"gemini-2.0-flash-thinking-exp-01-21\": 1.25,\n  \"gemini-2.0-flash-lite-preview-02-05\": 0.0375,\n  \"gemini-2.5-flash-lite-preview-06-17\": 0.75,\n  \"gemini-2.0-flash-lite-preview\": 0.0375,\n  \"gemini-2.0-pro-exp\": 1.25,\n  \"gemini-2.0-pro-exp-02-05\": 1.25,\n  \"gemini-2.0-flash-exp-image-generation\": 1.25,\n  \"gemini-2.0-flash-preview-image-generation\": 1.25,\n  \"gemini-2.5-flash-image-preview\": 0.15,\n  \"gemini-embedding-exp-03-07\": 0.065,\n  \"gemini-embedding-exp\": 0.065,\n  \"embedding-001\": 0.065,\n  \"gemini-2.0-flash-lite\":  0.35,\n  \"gemini-2.0-flash-lite-001\": 0.35,\n  \"gemma-3-27b-it\": 1.25,\n  \"gemini-2.5-pro-preview-tts\": 0.5,\n  \"gemini-2.5-pro-exp-03-25\": 1.25,\n  \"gemini-2.5-pro-preview-03-25\": 1.25,\n  \"gemini-2.5-pro-preview-05-06\": 1.25,\n  \"gemini-2.5-pro-preview-06-05\": 1.25,\n  \"gemini-2.5-pro\": 1.25,\n  \"gemini-2.5-flash-preview-04-17\": 0.075,\n  \"gemini-2.5-flash-preview-04-17-thinking\": 0.075,\n  \"gemini-2.5-flash-preview-05-20\": 0.15,\n  \"gemini-2.5-flash-preview-tts\": 0.25,\n  \"gemini-2.5-flash\": 0.325,\n  \"imagen-3.0-generate-002\": 0.075,\n	\"imagen-4.0-generate-preview-06-06\": 0.02,\n	\"imagen-4.0-ultra-generate-preview-06-06\": 0.03,\n  \"veo-2.0-generate-001\": 0.175,\n  \"claude-instant-1.2\": 0.4,\n  \"claude-2.0\": 4,\n  \"claude-2.1\": 4,\n  \"claude-3-haiku-20240307\": 0.13,\n  \"claude-3-5-haiku-20241022\": 0.65,\n  \"claude-3-sonnet-20240229\": 1.55,\n  \"claude-3-opus-20240229\": 7.55,\n  \"claude-3-5-sonnet-20240620\": 1.55,\n  \"claude-3-5-sonnet-20241022\": 1.55,\n  \"claude-3-5-sonnet-latest\": 1.55,\n  \"claude-3-7-sonnet-20250219\": 1.55,\n  \"claude-3-7-sonnet-latest\": 1.55,\n  \"claude-sonnet-4-20250514\": 1.55,\n  \"claude-opus-4-20250514\": 7.55,\n  \"claude-opus-4-1-20250805\": 7.55,\n  \"txt2img\": 10,\n  \"img2img\": 10,\n  \"txt2video\": 220,\n  \"img2video\": 220,\n  \"deepseek-chat\": 0.28,\n  \"deepseek-coder\": 0.28,\n  \"deepseek-reasoner\": 0.28,\n  \"qwen-1.8b-chat\": 1.4286,\n  \"qwen-1.8b-longcontext-chat\": 1.4286,\n  \"qwen-14b-chat\": 1.4286,\n  \"qwen-72b-chat\": 1.4286,\n  \"qwen-7b-chat\": 1.4286,\n  \"qwen-audio-chat\": 1.4286,\n  \"qwen-audio-turbo\": 1.4286,\n  \"qwen-coder-plus\": 1.4286,\n  \"qwen-coder-plus-latest\": 1.4286,\n  \"qwen-coder-turbo\": 1.4286,\n  \"qwen-coder-turbo-latest\": 1.4286,\n  \"qwen-math-plus\": 1.4286,\n  \"qwen-math-plus-latest\": 1.4286,\n  \"qwen-math-turbo\": 1.4286,\n  \"qwen-math-turbo-latest\": 1.4286,\n  \"qwen-max\": 1.2,\n  \"qwen-max-latest\": 1.2,\n  \"qwen-max-longcontext\": 1.4286,\n  \"qwen-plus\": 0.6,\n  \"qwen-plus-latest\": 0.6,\n  \"qwen-turbo\": 0.08,\n  \"qwen-turbo-latest\": 0.08,\n  \"qwen-vl-chat-v1\": 1.4286,\n  \"qwen-vl-max\": 1.4286,\n  \"qwen-vl-max-latest\": 1.4286,\n  \"qwen-vl-ocr\": 1.4286,\n  \"qwen-vl-ocr-latest\": 1.4286,\n  \"qwen-vl-plus\": 1.4286,\n  \"qwen-vl-plus-latest\": 1.4286,\n  \"qwen-vl-v1\": 1.4286,\n  \"qwen1.5-0.5b-chat\": 1.4286,\n  \"qwen1.5-1.8b-chat\": 1.4286,\n  \"qwen1.5-110b-chat\": 1.4286,\n  \"qwen1.5-14b-chat\": 1.4286,\n  \"qwen1.5-32b-chat\": 1.4286,\n  \"qwen1.5-72b-chat\": 1.4286,\n  \"qwen1.5-7b-chat\": 1.4286,\n  \"qwen2-0.5b-instruct\": 1.4286,\n  \"qwen2-1.5b-instruct\": 1.4286,\n  \"qwen2-57b-a14b-instruct\": 1.4286,\n  \"qwen2-72b-instruct\": 1.4286,\n  \"qwen2-7b-instruct\": 1.4286,\n  \"qwen2-audio-instruct\": 1.4286,\n  \"qwen2-math-1.5b-instruct\": 1.4286,\n  \"qwen2-math-72b-instruct\": 1.4286,\n  \"qwen2-math-7b-instruct\": 1.4286,\n  \"qwen2-vl-2b-instruct\": 1.4286,\n  \"qwen2-vl-7b-instruct\": 1.4286,\n  \"qwen2.5-0.5b-instruct\": 1.4286,\n  \"qwen2.5-1.5b-instruct\": 1.4286,\n  \"qwen2.5-14b-instruct\": 1.4286,\n  \"qwen2.5-32b-instruct\": 1.4286,\n  \"qwen2.5-3b-instruct\": 1.4286,\n  \"qwen2.5-72b-instruct\": 1.4286,\n  \"qwen2.5-7b-instruct\": 1.4286,\n  \"qwen2.5-coder-0.5b-instruct\": 1.4286,\n  \"qwen2.5-coder-1.5b-instruct\": 1.4286,\n  \"qwen2.5-coder-14b-instruct\": 1.4286,\n  \"qwen2.5-coder-32b-instruct\": 1.4286,\n  \"qwen2.5-coder-3b-instruct\": 1.4286,\n  \"qwen2.5-coder-7b-instruct\": 1.4286,\n  \"qwen2.5-math-1.5b-instruct\": 1.4286,\n  \"qwen2.5-math-72b-instruct\": 1.4286,\n  \"qwen2.5-math-7b-instruct\": 1.4286,\n  \"qwen-max-2025-01-25\": 1.4286,\n  \"text-embedding-v1\": 0.05,\n  \"text-embedding-v2\": 0.05,\n  \"text-embedding-v3\": 0.05,\n  \"text-embedding-async-v1\": 0.05,\n  \"text-embedding-async-v2\": 0.05,\n  \"ali-stable-diffusion-v1.5\": 8,\n  \"ali-stable-diffusion-xl\": 8,\n  \"wanx-v1\": 8\n}'),
('Notice','请加群关注最新公告获取最新资讯！'),
('passkey.attachment_preference',''),
('passkey.origins',''),
('passkey.rp_display_name','AIGuoGuo'),
('passkey.rp_id','https://api.aiguoguo199.com'),
('passkey.user_verification','preferred'),
('PasswordRegisterEnabled','false'),
('PoolMode','polling'),
('PreConsumedQuota','5000'),
('QuotaForInvitee','500000'),
('QuotaForInviter','500000'),
('RegisterEnabled','false'),
('RetryTimes','2'),
('SelfUseModeEnabled','false'),
('ServerAddress','https://api.aiguoguo199.com'),
('SMTPAccount','admin@aichat199.com'),
('SMTPFrom','admin@aichat199.com'),
('SMTPPort','465'),
('SMTPServer','mail.aichat199.com'),
('SMTPToken','Admin23_Ex03oIi2'),
('StatisticsInfoEnabled','true'),
('SystemName','果果API'),
('TopUpLink','https://shop.aichat199.com/');

/*Table structure for table `orders` */

DROP TABLE IF EXISTS `orders`;

CREATE TABLE `orders` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `order_no` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `card_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `completed_at` bigint(20) DEFAULT NULL,
  `note` text COLLATE utf8mb4_unicode_ci,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_orders_order_no` (`order_no`),
  KEY `idx_orders_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

/*Data for the table `orders` */

/*Table structure for table `system_config` */

DROP TABLE IF EXISTS `system_config`;

CREATE TABLE `system_config` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `key` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` text COLLATE utf8mb4_unicode_ci,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_system_config_key` (`key`),
  KEY `idx_system_config_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

/*Data for the table `system_config` */

insert  into `system_config`(`id`,`key`,`value`,`created_at`,`updated_at`,`deleted_at`) values 
(1,'db_host','192.168.0.113','2025-11-29 06:51:13.590','2025-11-29 06:51:13.590',NULL),
(2,'db_port','3316','2025-11-29 06:51:13.591','2025-11-29 06:51:13.591',NULL),
(3,'db_name','guo_depot','2025-11-29 06:51:13.592','2025-11-29 06:51:13.592',NULL),
(4,'db_user','root','2025-11-29 06:51:13.592','2025-11-29 06:51:13.592',NULL),
(5,'system_initialized','true','2025-11-29 06:51:13.593','2025-11-29 06:51:13.593',NULL),
(6,'site_name','ChatGPT充值平台','2025-11-29 06:51:13.593','2025-11-29 06:51:13.593',NULL);

/*Table structure for table `users` */

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password` longtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `display_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `role` bigint(20) DEFAULT '1',
  `status` bigint(20) DEFAULT '1',
  `email` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `github_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `wechat_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `lark_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `oidc_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `access_token` char(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `quota` bigint(20) DEFAULT '0',
  `used_quota` bigint(20) DEFAULT '0',
  `request_count` bigint(20) DEFAULT '0',
  `group` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT 'default',
  `aff_code` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `inviter_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_access_token` (`access_token`),
  UNIQUE KEY `idx_users_aff_code` (`aff_code`),
  UNIQUE KEY `uni_users_username` (`username`),
  KEY `idx_users_username` (`username`),
  KEY `idx_users_display_name` (`display_name`),
  KEY `idx_users_email` (`email`),
  KEY `idx_users_git_hub_id` (`github_id`),
  KEY `idx_users_we_chat_id` (`wechat_id`),
  KEY `idx_users_lark_id` (`lark_id`),
  KEY `idx_users_oidc_id` (`oidc_id`),
  KEY `idx_users_inviter_id` (`inviter_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

/*Data for the table `users` */

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
