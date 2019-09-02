import { AppState } from '../index';
import { connect } from 'react-redux';
import { AuthRoute } from '../components/Route/AuthRoute';
import { fetchSettings } from '../actions/settingsAction';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';

const mapStateToProps = (state: AppState) => ({
  isLoggedIn: !!state.authStatus.userId,
  loading: !state.authStatus.checked,
  alreadyLoaded: state.authStatus.checked,
  error: state.error.errorMessage,
});
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  load: () => {
    dispatch(fetchSettings());
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(AuthRoute);
