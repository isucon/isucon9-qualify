import { connect } from "react-redux";
import { AppState } from "../index";
import ItemListPage from "../pages/ItemListPage";
import { mockItems } from "../mocks";

const mapStateToProps = (state: AppState) => ({
  items: mockItems, // TODO
  errorType: state.error.errorType,
  loading: false // TODO state.page.isLoading,
});
const mapDispatchToProps = (dispatch: any) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(ItemListPage);
