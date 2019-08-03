import React, {ReactNode} from 'react';

import { Container, MuiThemeProvider, createMuiTheme } from '@material-ui/core';
import {ErrorType} from "../reducers/errorReducer";
import NotFoundPage from "../pages/error/NotFoundPage";
import InternalServerErrorPage from "../pages/error/InternalServerErrorPage";

const themeInstance = createMuiTheme({
    palette: {
        background: {
            default: 'white'
        },
    },
});

type Props = {
    errorType: ErrorType
    children: ReactNode
}

class BasePageComponent extends React.Component<Props> {
    private getProperComponent() {
        const error: ErrorType = this.props.errorType;

        switch (error) {
            case "NO_ERROR":
                return this.props.children;
            case "NOT_FOUND":
                return NotFoundPage;
            case "INTERNAL_SERVER_ERROR":
                return InternalServerErrorPage;
            default:
                return InternalServerErrorPage;
        }
    }

    render() {
        const contentComponent = this.getProperComponent();

        return (
            <MuiThemeProvider theme={themeInstance}>
                <Container maxWidth="lg">
                    {contentComponent}
                </Container>
            </MuiThemeProvider>
        );
    }
}

export { BasePageComponent }