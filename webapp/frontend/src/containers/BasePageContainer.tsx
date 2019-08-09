import { connect } from "react-redux";
import { AppState } from "../index";
import { fetchSettings } from "../actions/settingsAction";
import { BasePageComponent } from "../components/BasePageComponent";

const mapStateToProps = (state: AppState) => ({
  loading: state.page.isLoading,
  alreadyLoaded: state.authStatus.checked
});
const mapDispatchToProps = (dispatch: any) => ({
  load: () => {
    dispatch(fetchSettings());
  }
});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(BasePageComponent);
