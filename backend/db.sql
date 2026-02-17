/*
SQLyog Ultimate v12.5.0 (64 bit)
MySQL - 5.7.16-log : Database - guo_depot
*********************************************************************
*/

/*!40101 SET NAMES utf8mb4 */;
/*!40101 SET CHARACTER_SET_CLIENT=utf8mb4 */;
/*!40101 SET CHARACTER_SET_RESULTS=utf8mb4 */;

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

/*
  注意：卡密表不在此处创建
  卡密表通过"菜单管理 -> 新增卡密菜单"功能动态创建
  表名格式：cards_{category}，例如：cards_cursor、cards_chatgpt 等
  表结构定义在 backend/model/menu.go 的 CreateCardTable 函数中
*/

/*Table structure for table `options` */

DROP TABLE IF EXISTS `options`;

CREATE TABLE `options` (
  `key` varchar(191) NOT NULL,
  `value` longtext,
  PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

/*Data for the table `options` */

insert  into `options`(`key`,`value`) values 
('Logo','https://img.alicdn.com/imgextra/i2/3566933417/O1CN01WDyh0I1b72BCVb6Co_!!3566933417.png_120x120.jpg'),
('SMTPAccount','admin@aichat199.com'),
('SMTPFrom','admin@aichat199.com'),
('SMTPPort','465'),
('SMTPServer','mail.aichat199.com'),
('SMTPToken','Admin23_Ex03oIi2'),

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

/*Table structure for table `users` */

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password` longtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `display_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` bigint(20) DEFAULT '1',
  `email` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `github_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `wechat_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `lark_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `oidc_id` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `access_token` char(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
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

/*Table structure for table `menus` */

DROP TABLE IF EXISTS `menus`;

CREATE TABLE `menus` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `parent_id` bigint(20) DEFAULT '0' COMMENT '父菜单ID, 0为顶级菜单',
  `title` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单标题',
  `key` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单唯一key',
  `path` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '路由路径',
  `icon` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '菜单图标(emoji)',
  `sort` bigint(20) DEFAULT '0' COMMENT '排序权重，数值越小越靠前',
  `status` tinyint(1) DEFAULT '1' COMMENT '状态: 1启用 0禁用',
  `is_delete` tinyint(1) DEFAULT '-1' COMMENT '是否删除: 1是 -1否',
  `deleted_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_menus_key` (`key`),
  KEY `idx_menus_parent_id` (`parent_id`),
  KEY `idx_menus_is_delete` (`is_delete`),
  KEY `idx_menus_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

/*Data for the table `menus` - 初始化数据 */

-- 使用变量方式处理父子关系，避免 MySQL 的 "You can't specify target table for update in FROM clause" 错误

-- 插入顶级菜单：控制台
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) 
VALUES (0,'控制台','dashboard','/admin/dashboard','📊',1,1,-1,NOW(),NOW());

-- 插入顶级菜单：用户管理
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) 
VALUES (0,'用户管理','user',NULL,'👥',2,1,-1,NOW(),NOW());
SET @user_menu_id = LAST_INSERT_ID();
-- 用户管理的子菜单
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) VALUES 
(@user_menu_id,'用户列表','users','/admin/users',NULL,1,1,-1,NOW(),NOW()),
(@user_menu_id,'角色管理','roles','/admin/roles',NULL,2,1,-1,NOW(),NOW());

-- 插入顶级菜单：订单管理
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) 
VALUES (0,'订单管理','order',NULL,'📦',3,1,-1,NOW(),NOW());
SET @order_menu_id = LAST_INSERT_ID();
-- 订单管理的子菜单
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) VALUES 
(@order_menu_id,'订单列表','orders','/admin/orders',NULL,1,1,-1,NOW(),NOW()),
(@order_menu_id,'退款管理','refunds','/admin/refunds',NULL,2,1,-1,NOW(),NOW());

-- 插入顶级菜单：镜像管理
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) 
VALUES (0,'镜像管理','mirror',NULL,'🔐',5,1,-1,NOW(),NOW());
SET @mirror_menu_id = LAST_INSERT_ID();
-- 镜像管理的子菜单
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) VALUES 
(@mirror_menu_id,'卡密管理','mirror-cards','/admin/mirror-cards',NULL,1,1,-1,NOW(),NOW());

-- 插入顶级菜单：系统设置
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) 
VALUES (0,'系统设置','system',NULL,'⚙️',6,1,-1,NOW(),NOW());
SET @system_menu_id = LAST_INSERT_ID();
-- 系统设置的子菜单
INSERT INTO `menus`(`parent_id`,`title`,`key`,`path`,`icon`,`sort`,`status`,`is_delete`,`created_at`,`updated_at`) VALUES 
(@system_menu_id,'基础设置','settings','/admin/settings',NULL,1,1,-1,NOW(),NOW()),
(@system_menu_id,'操作日志','logs','/admin/logs',NULL,2,1,-1,NOW(),NOW()),
(@system_menu_id,'菜单管理','menu-management','/admin/menu-management',NULL,3,1,-1,NOW(),NOW());

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
