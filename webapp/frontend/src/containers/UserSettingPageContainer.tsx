import { connect } from 'react-redux';
import { AppState } from '../index';
import UserSettingPage from '../pages/UserSettingPage';
import { Dispatch } from 'redux';

const mapStateToProps = (state: AppState) => ({
  id: state.authStatus.userId,
  accountName: state.authStatus.accountName,
  address: state.authStatus.address,
  numSellItems: state.authStatus.numSellItems,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(UserSettingPage);
