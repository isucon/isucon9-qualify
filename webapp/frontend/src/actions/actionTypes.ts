import { Action } from 'redux';
import { AuthActions } from './authenticationActions';
import { BuyActions } from './buyAction';
import { ErrorActions } from './errorAction';
import { FetchTimelineActions } from './fetchTimelineAction';
import { FetchTransactionActions } from './fetchTransactionsAction';
import { FetchUserItemsActions } from './fetchUserItemsAction';
import { LocationChangeActions } from './locationChangeAction';
import { PostBumpActions } from './postBumpAction';
import { FetchUserPageDataActions } from './fetchUserPageDataAction';
import { PostCompleteActions } from './postCompleteAction';
import { PostItemEditActions } from './postItemEditAction';
import { PostShippedActions } from './postShippedAction';
import { PostShippedDoneActions } from './postShippedDoneAction';
import { RegisterActions } from './registerAction';
import { SellingItemActions } from './sellingItemAction';
import { SettingsActions } from './settingsAction';
import { FetchItemActions } from './fetchItemAction';
import { RouterAction } from 'connected-react-router';
import { SnackBarActions } from './snackBarAction';
import { SnackBarVariant } from '../components/SnackBar';

type LibraryActions = RouterAction;

export type ActionTypes =
  | LibraryActions
  | AuthActions
  | BuyActions
  | ErrorActions
  | FetchItemActions
  | FetchTimelineActions
  | FetchTransactionActions
  | FetchUserItemsActions
  | FetchUserPageDataActions
  | LocationChangeActions
  | PostBumpActions
  | PostCompleteActions
  | PostItemEditActions
  | PostShippedActions
  | PostShippedDoneActions
  | RegisterActions
  | SellingItemActions
  | SettingsActions
  | SnackBarActions;

export interface SnackBarAction<T> extends Action<T> {
  snackBarMessage: string;
  variant: SnackBarVariant;
}
