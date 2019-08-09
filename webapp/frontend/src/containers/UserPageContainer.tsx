import { connect } from 'react-redux';
import { AppState } from '../index';
import { mockItems, mockUser } from '../mocks';
import UserPage from '../pages/UserPage';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';

const mapStateToProps = (state: AppState) => ({
  items: mockItems, // TODO
  user: mockUser,
  errorType: state.error.errorType,
  loading: false, // TODO state.page.isLoading,
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(UserPage);
