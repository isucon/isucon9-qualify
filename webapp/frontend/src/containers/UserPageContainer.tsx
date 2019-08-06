import {connect} from "react-redux";
import {AppState} from "../index";
import {mockItems, mockUser} from "../mocks";
import UserPage from "../pages/UserPage";

const mapStateToProps = (state: AppState) => ({
    items: mockItems, // TODO
    user: mockUser,
    errorType: state.error.errorType,
    loading: false,// TODO state.page.isLoading,
});
const mapDispatchToProps = (dispatch: any) => ({
});

export default connect(mapStateToProps, mapDispatchToProps)(UserPage);
