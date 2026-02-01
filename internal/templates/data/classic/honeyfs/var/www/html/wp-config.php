<?php
/**
 * The base configuration for WordPress
 */
define( 'DB_NAME', 'wp_prod_db' );
define( 'DB_USER', 'wp_user' );
define( 'DB_PASSWORD', 'SuperSecurePass2024!' ); // <--- LEURRE ICI
define( 'DB_HOST', 'localhost' );
define( 'DB_CHARSET', 'utf8' );
define( 'DB_COLLATE', '' );

/** Authentication Unique Keys and Salts. */
define( 'AUTH_KEY',         'put your unique phrase here' );
define( 'SECURE_AUTH_KEY',  'put your unique phrase here' );