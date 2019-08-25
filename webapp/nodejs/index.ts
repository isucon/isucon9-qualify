import express from "express";
// @ts-ignore
import mysql from "mysql2";


const connection = mysql.createConnection({
  host: 'localhost',
  user: 'isucari',
  password: 'isucari',
  database: 'isucari',
});

console.log("connected", connection);

process.exit();
