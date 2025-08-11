import mysql from 'mysql2/promise';

export type MySQLResultRows = mysql.RowDataPacket[];
export type MySQLResultSetHeader = mysql.ResultSetHeader;

let pool: mysql.Pool;

export function initDB() {
  pool = mysql.createPool({
    host: process.env['MYSQL_HOST'] || '127.0.0.1',
    port: parseInt(process.env['MYSQL_PORT'] || '3306'),
    user: process.env['MYSQL_USER'] || 'isucari',
    password: process.env['MYSQL_PASS'] || 'isucari',
    database: process.env['MYSQL_DBNAME'] || 'isucari',
    connectionLimit: 100,
    timezone: '+00:00',
    namedPlaceholders: false,
    waitForConnections: true,
    queueLimit: 0,
  });
}

export async function getConnection(): Promise<mysql.PoolConnection> {
  return await pool.getConnection();
}
