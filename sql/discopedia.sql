--
-- Table structure for table `login`
--

DROP TABLE IF EXISTS `login`;
CREATE TABLE `login` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `email` varchar(50) NOT NULL,
  `password` char(128) NOT NULL,
  `unique_identifier` char(9) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `session_code` (`unique_identifier`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `login`
--

LOCK TABLES `login` WRITE;
UNLOCK TABLES;

--
-- Table structure for table `logs`
--

DROP TABLE IF EXISTS `logs`;
CREATE TABLE `logs` (
  `host_id` int(11) NOT NULL,
  `guest_id` int(11) NOT NULL,
  `date` datetime NOT NULL,
  KEY `host_id` (`host_id`),
  KEY `guest_id` (`guest_id`),
  CONSTRAINT `logs_ibfk_1` FOREIGN KEY (`host_id`) REFERENCES `login` (`id`),
  CONSTRAINT `logs_ibfk_2` FOREIGN KEY (`guest_id`) REFERENCES `login` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `logs`
--

LOCK TABLES `logs` WRITE;
UNLOCK TABLES;

--
-- Table structure for table `tierlist`
--

DROP TABLE IF EXISTS `tierlist`;
CREATE TABLE `tierlist` (
  `user_id` int(11) NOT NULL,
  `name` text NOT NULL,
  `artist` varchar(80) NOT NULL,
  `year` smallint(5) unsigned NOT NULL,
  `tier` char(1) DEFAULT NULL,
  `image` varchar(80) DEFAULT NULL,
  KEY `user_id` (`user_id`),
  CONSTRAINT `tierlist_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `login` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `tierlist`
--

LOCK TABLES `tierlist` WRITE;
UNLOCK TABLES;