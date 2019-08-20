import { AppState } from '../index';
import { Dispatch } from 'redux';
import { connect } from 'react-redux';
import { push } from 'connected-react-router';
import { routes } from '../routes/Route';
import { Header } from '../components/Header';

const mapStateToProps = (state: AppState) => ({
  isLoggedIn: !!state.authStatus.userId,
  ownUserId: state.authStatus.userId || 0,
});

const mapDispatchToProps = (dispatch: Dispatch) => ({
  goToTopPage: () => {
    dispatch(push(routes.timeline.path));
  },
  goToUserPage: (userId: number) => {
    dispatch(push(routes.user.getPath(userId)));
  },
  goToSettingPage: () => {
    dispatch(push(routes.userSetting.path));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(Header);
