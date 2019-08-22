import { Action } from 'redux';

export const PATH_NAME_CHANGE = 'PATH_NAME_CHANGE';

export interface PathNameChangeAction extends Action<typeof PATH_NAME_CHANGE> {}

export const pathNameChangeAction = (): PathNameChangeAction => ({
  type: PATH_NAME_CHANGE,
});
