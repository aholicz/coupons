DROP TABLE IF EXISTS `coupon_code`;

CREATE TABLE `coupon_code` (
  `code_id` bigint NOT NULL AUTO_INCREMENT,
  `coupon` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  PRIMARY KEY (`code_id`),
  UNIQUE KEY `coupon` (`coupon`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


