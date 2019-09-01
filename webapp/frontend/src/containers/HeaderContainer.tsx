import { AppState } from '../index';
import { Dispatch } from 'redux';
import { connect } from 'react-redux';
import { push } from 'connected-react-router';
import { routes } from '../routes/Route';
import { Header } from '../components/Header';
import { CategorySimple } from '../dataObjects/category';

const mapStateToProps = (state: AppState) => ({
  isLoggedIn: !!state.authStatus.userId,
  ownUserId: state.authStatus.userId || 0,
  // Note: Showing only parent category
  categories: state.categories.categories.filter(
    (category: CategorySimple) => category.parentId === 0,
  ),
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
  goToCategoryItemList: (categoryId: number) => {
    dispatch(push(routes.categoryTimeline.getPath(categoryId)));
  },
  onClickTitle: (isLoggedIn: boolean) => {
    if (isLoggedIn) {
      dispatch(push(routes.timeline.path));
      return;
    }
    dispatch(push(routes.top.path));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(Header);
