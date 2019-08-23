import { connect } from 'react-redux';
import ItemBuyFormComponent from '../components/ItemBuyFormComponent';
import { buyItemAction } from '../actions/buyAction';
import { AppState } from '../index';
import { ThunkDispatch } from 'redux-thunk';
import { ActionTypes } from '../actions/actionTypes';

const mapStateToProps = (state: AppState) => ({
  item: state.viewingItem.item,
  errors: state.formError.buyFormError,
  loadingBuy: state.buyPage.loadingBuy,
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, ActionTypes>,
) => ({
  onBuyAction: (itemId: number, cardNumber: string) => {
    dispatch(buyItemAction(itemId, cardNumber));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ItemBuyFormComponent);
