import logger from './logger';
import checkLoginChange from './checkLocationChange';
import { Middleware } from 'redux';

const middleware: Middleware[] = [logger, checkLoginChange];

export default middleware;
