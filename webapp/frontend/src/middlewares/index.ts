import logger from './logger';
import checkLocationChange from './checkLocationChange';
import { Middleware } from 'redux';

const middleware: Middleware[] = [logger, checkLocationChange];

export default middleware;
