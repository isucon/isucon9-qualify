import { AppState } from '../index';
import { AnyAction } from 'redux';
import { connect } from 'react-redux';
import { NonAuthRoute } from '../components/Route/NonAuthRoute';
import { fetchSettings } from '../actions/settingsAction';
import { ThunkDispatch } from 'redux-thunk';

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
)(NonAuthRoute);
