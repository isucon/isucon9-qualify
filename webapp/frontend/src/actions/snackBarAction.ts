import { Action } from 'redux';

export const SNACK_BAR_CLOSE = 'SNACK_BAR_CLOSE';

export type SnackBarActions = SnackBarClose;

export interface SnackBarClose extends Action<typeof SNACK_BAR_CLOSE> {}

export const closeSnackBarAction = (): SnackBarClose => ({
  type: SNACK_BAR_CLOSE,
});
