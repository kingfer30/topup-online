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

DROP TABLE IF EXISTS `sales_talks`;
CREATE TABLE `sales_talks` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `title` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单标题',
  `sort` int(11) DEFAULT '0' COMMENT '排序权重，数值越小越靠前',
  `zh_content` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '中文内容',
  `en_content` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '英文内容',
  `ru_content` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '俄文内容',
  `tag`varchar(100) DEFAULT NULL COMMENT '标签',
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_tag` (`tag`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- insert into `sales_talks` (`title`, `sort`, `zh_content`, `en_content`, `ru_content`, `created_at`, `tag`) values
-- ('订单完成 - 通用','0','如果您喜欢我们的服务，请在订单模块给我们好评，您将获得一张价值相当于订单总金额 5% 的礼品卡?️\n\n关注我们的频道，获取更多优惠：https://t.me/AI_GUO_GUO\n\n祝您愉快！','If you liked our service, you can leave us a good review in the order module, you will receive a gift card worth 5% of the total amount paid?️\n\nFollow our channel to get more discounts: https://t.me/AI_GUO_GUO\n\nHave a good day) ','Ваш заказ выполнен !\n\nЕсли вам понравился наш сервис, вы можете оставить нам хороший отзыв в модуле заказа, вы получите подарочную карту на сумму 5% от общей оплаченной суммы?️\nПодписывайтесь на наш канал, чтобы получать больше выгодных предложений: https://t.me/AI_GUO_GUO\n\nхорошего дня )','2026-02-21 01:42:50.099','通用'),
-- ('订单完成 - 快捷发货用','1','您的订单已完成！\n\n我们的配送速度堪比闪电麦昆，服务精准无误；如果您满意，请留下好评，即可立即获得价值订单总额 5% 的礼品卡。?️\n\n订阅我们的频道，获取更多优惠信息：https://t.me/AI_GUO_GUO\n\n祝您愉快！','Your order has been completed!\n\nOur delivery speed is as fast as Lightning McQueen; our service is as precise as the periodic table; if you agree, please leave a positive review on your order, and you\'ll immediately receive a gift card worth 5% of the total order amount. ?️\n\nSubscribe to our channel to receive more great deals: https://t.me/AI_GUO_GUO\n\nHave a nice day!','Ваш заказ выполнен !\n\nСкорость нашей доставки быстра, как Молния Маккуин; сервис точен, как периодическая таблица Менделеева; — если вы согласны с этим, пожалуйста, оставьте положительный отзыв в заказе, и вы сразу же получите подарочную карту на сумму, равную 5% от общей суммы заказа.?️\n\nПодписывайтесь на наш канал, чтобы получать больше выгодных предложений: https://t.me/AI_GUO_GUO\n\nхорошего дня )','2026-02-21 01:43:59.295','通用'),
-- ('发送订单唯一编码','0','您好先生，请将唯一的订单代码发送给我。','Hello sir, please send me the unique order code','Здравствуйте, сэр, пожалуйста, пришлите мне уникальный код заказа','2026-02-21 01:44:38.259','通用'),
-- ('发送邮件验证码','3','请将您邮箱中的验证码发送给我。','Please send me the confirmation code from your email.','Пожалуйста, отправьте мне код подтверждения из почты.','2026-02-21 01:45:14.554','通用'),
-- ('购买补差额商品','0','https://plati.market/itm/3785171\n\n请购买此商品并把订单号发给我。','https://plati.market/itm/3785171\n\nPlease purchase one of these and send me the unique order code.','https://plati.market/itm/3785171\n\nПожалуйста, приобретите один из них и пришлите мне уникальный код заказа','2026-02-21 01:46:05.495','通用'),
-- ('被发现发的错误日期','0','非常抱歉，我们的同事发送的日期略早一些。\n\n针对此问题，我们将提供以下补偿：\n\n20/30*3=$2\n\n我们将以礼品卡的形式退还您 2 美元。您是否同意？\n\n您不必过于担心，因为即使您的 Cursor 订阅到期，它仍然可以继续使用大约一周。','We\'re very sorry, but unfortunately our colleague sent a slightly earlier date.\n\nIn response to this issue, we\'re providing the following compensation:\n\n20/30*3=$2\n\nWe\'ll refund you $2 via a gift card. Do you approve this?\n\nYou don\'t need to worry too much, because after your Cursor subscription expires, it can still work for about a week.','Очень сожалею, но, к сожалению, наш коллега прислал немного более раннюю дату\n\nВ ответ на эту проблему мы предоставляем следующую компенсацию\n\n20/30*3=2$\n\nМы вернем вам 2$ с помощью подарочной карты. Вы это одобряете?\n\nВам также не нужно слишком беспокоиться, потому что после истечения срока подписки Cursor он все еще может работать около недели.\n','2026-02-21 01:46:42.440','通用'),
-- ('封号赔偿方案','0','已使用 21 天\n\n我们提供 30 天保修，根据订单金额，计算公式如下：\n\n22 / 30 * 21 = 15.4 美元\n\n22 / 30 * (30 - 21) = 6.6 美元\n\n您希望选择哪种付款方式？\n\n1. 您需要支付 15.4 美元的差价，我们将为您开通新账​​户。\n\n2. 我们将以礼品卡的形式退还您 6.6 美元。','Used for 21 days\n\nWe offer a 30-day warranty, and the calculation formula, depending on the order value, is as follows:\n\n22 / 30 * 21 = $15.4\n22 / 30 * (30 - 21) = $6.6\n\nWhich payment method would you prefer?\n1. You\'ll need to pay the difference of $15.4, and we\'ll send you a new account.\n2. We\'ll refund you $6.6 as a gift card.','Использовался в течение 21 дней\n\nМы предоставляем гарантийный срок в 30 дней, а формула расчета в зависимости от стоимости заказа выглядит следующим образом\n\n22 / 30 * 21 =15.4$\n22 / 30 * (30 - 21) =6.6$\n\nКакой метод вы бы предпочли принять?\n1. Вам нужно доплатить разницу в 15.4$, мы отправим вам новый аккаунт\n2. Мы вернем вам 6.6$ в виде подарочной карты','2026-02-21 01:47:24.786','通用'),
-- ('使用礼品卡退款','0','14$\n我会把礼品卡发送到这个账户:  sinusoidoss@gmail.com','$14\nI\'ll send the gift card to this account: sinusoidoss@gmail.com','14$\nЯ отправлю подарочную карту на этот счет:  sinusoidoss@gmail.com','2026-02-21 01:48:05.545','通用'),
-- ('gpt-付款链接非USD','0','您好，您订单的支付货币为欧元，因此订单处理费将会增加。\n\n请您先退出当前账户，然后使用美国IP地址重新登录，再次获取支付链接并发送给我。','Hello, your payment currency for the linked order is Euro, which will result in an increased order processing fee.\nPlease log out and log in from a US IP address, retrieve the payment link again, and send it to me.','Здравствуйте, валюта оплаты вашего заказа по ссылке - евро, что приведет к увеличению комиссии за обработку заказа\nПожалуйста, выйдите из системы и войдите в систему с IP-адреса в Соединенных Штатах, снова получите ссылку для оплаты и отправьте ее мне\n','2026-02-21 01:49:43.790','GPT'),
-- ('gpt-判断付款链接是否正确','0','首先，您可以获取支付链接并验证其有效性。\n\n如果链接格式为 pay.openai.com/xx，则此链接有效。\n\n如果链接格式为 chatgpt.com/xx，则此链接无效。','First, you can get the payment link and show me if it\'s valid.\nIf it\'s like this: pay.openai.com/xx , then this link is valid.\nIf it\'s like this: chatgpt.com/xx , then this link is not valid.','Сначала вы можете получить ссылку для оплаты и показать мне, нормальная ли она.\nЕсли это так: pay.openai.com/xx , то эта ссылка нормальная\nЕсли это так: chatgpt.com/xx , то эта ссылка не является обычной','2026-02-21 01:50:39.570','GPT'),
-- ('gpt-引导生成session','0','您可以通过其 API 获取。\n\n1. 在聊天窗口中，打开 chatgpt.com 页面，然后按键盘上的 F12 键。\n\n2. 选择“控制台”，输入以下代码，然后按回车键。\n\nfetch(\"/api/auth/session\").then(r => r.json()).then(({ accessToken }) => {\nfetch(\"/backend-api/payments/checkout\", {\n\"method\": \"POST\",\n\n\"headers\": { \"authorization\": `Bearer ${accessToken}`, },\n\n}).then(r => r.json()).then(d => window.open(d.url))\n\n})\n\n3. 稍等片刻，将打开一个包含正确支付链接的新窗口。复制该链接并发送给我。\n\n4. 如果粘贴后没有收到回复，请按照说明输入“允许粘贴”，然后按回车键，即可再次粘贴。','You can get it through its API.\n\n1. In your chat. On the chatgpt.com page, press F12 on your keyboard.\n2. Select \"Console\", enter the following code, and press enter.\n\nfetch(\"/api/auth/session\").then(r => r.json()).then(({ accessToken }) => {\nfetch(\"/backend-api/payments/checkout\", {\n\"method\": \"POST\",\n\"headers\": { \"authorization\": `Bearer ${accessToken}`, },\n}).then(r => r.json()).then(d => window.open(d.url))\n})\n\n3. After a short wait, it will open a new window with the correct payment link. Copy it and send it to me.\n4. If you don\'t receive a response after pasting, follow the instructions and type \"allow pasting\", press enter, then you can paste. line again.','Можно получить через его API-интерфейс\n\n1. В вашем чате.На странице chatgpt.com нажмите клавишу F12 на клавиатуре\n2. Выберите \"Console\", введите следующий код и нажмите enter\n\nfetch(\"/api/auth/session\").then(r => r.json()).then(({ accessToken }) => {\n  fetch(\"/backend-api/payments/checkout\", {\n    \"method\": \"POST\",\n    \"headers\": { \"authorization\": `Bearer ${accessToken}`, },\n  }).then(r => r.json()).then(d => window.open(d.url))\n})\n\n3. После некоторого ожидания он откроет новое окно, в котором будет указана правильная ссылка для оплаты, скопируйте и отправьте ее мне\n4. Если после вставки вы не получите ответа, следуйте инструкциям и введите \"allow pasting\", нажмите enter, после чего вы сможете вставить строку снова.\n','2026-02-21 01:51:45.175','GPT'),
-- ('gpt-取消连续订阅','0','头像 - 设置 - 订阅，取消当前的月度订阅计划，该计划在本月仍将有效。','Avatar - Settings - Subscription, cancel the ongoing monthly subscription plan, it will still be active for the current month','Аватар - Настройки - подписка, отмените план постоянной ежемесячной подписки, он по-прежнему будет действовать в течение текущего месяца\n','2026-02-21 01:52:53.147','GPT');

DROP TABLE IF EXISTS `digiseller_orders`;
CREATE TABLE `digiseller_orders` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `inv` bigint(20) NOT NULL COMMENT 'Digiseller 发票号，全局唯一',
  `unique_code` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户 16 位唯一码',
  `id_goods` bigint(20) DEFAULT NULL COMMENT '商品 ID',
  `amount` decimal(10,2) DEFAULT NULL COMMENT '实收金额',
  `type_curr` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收款货币类型（WMZ/WMR/WME/WMX）',
  `amount_usd` decimal(10,2) DEFAULT NULL COMMENT '等值 USD 金额',
  `profit` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '扣佣后净收益',
  `date_pay` datetime(3) DEFAULT NULL COMMENT '支付时间',
  `email` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '买家邮箱',
  `agent_id` int(11) DEFAULT NULL COMMENT '代理商 ID',
  `agent_percent` decimal(5,2) DEFAULT NULL COMMENT '代理商佣金比例',
  `cnt_goods` int(11) DEFAULT NULL COMMENT '购买数量（非固定价格商品）',
  `promo_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '买家使用的优惠码',
  `bonus_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '赠送给买家的优惠码',
  `cart_uid` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '购物车 UID',
  `uc_state` tinyint(4) DEFAULT NULL COMMENT '唯一码状态：1未验证 2已交付待确认 3已确认 4已驳回 5已验证未交付',
  `uc_date_check` datetime(3) DEFAULT NULL COMMENT '唯一码验证时间',
  `uc_date_delivery` datetime(3) DEFAULT NULL COMMENT '商品交付时间',
  `uc_date_confirmed` datetime(3) DEFAULT NULL COMMENT '交付确认时间',
  `uc_date_refuted` datetime(3) DEFAULT NULL COMMENT '交付驳回时间',
  `options_json` text COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '额外参数列表（JSON）',
  `created_at` datetime(3) DEFAULT NULL COMMENT '首次入库时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '最近更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_inv` (`inv`),
  KEY `idx_unique_code` (`unique_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

