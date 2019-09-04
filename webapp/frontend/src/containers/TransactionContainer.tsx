import { AppState } from '../index';
import { Dispatch } from 'redux';
import { connect } from 'react-redux';
import { push } from 'connected-react-router';
import { routes } from '../routes/Route';
import { TransactionComponent } from '../components/TransactionComponent';
import { TransactionItem } from '../dataObjects/item';

const mapStateToProps = (state: AppState) => ({});

const mapDispatchToProps = (dispatch: Dispatch) => ({
  onClickCard(item: TransactionItem) {
    if (item.status === 'on_sale') {
      dispatch(push(routes.item.getPath(item.id)));
      return;
    }

    dispatch(push(routes.transaction.getPath(item.id)));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(TransactionComponent);
