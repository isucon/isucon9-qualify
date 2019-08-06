import React, {ReactNode} from 'react';

import { Container, MuiThemeProvider, createMuiTheme } from '@material-ui/core';

const themeInstance = createMuiTheme({
    palette: {
        background: {
            default: 'white'
        },
    },
});

export type Props = {
    children: ReactNode
}

class BasePageComponent extends React.Component<Props> {
    render() {
        return (
            <MuiThemeProvider theme={themeInstance}>
                <Container maxWidth="lg">
                    {this.props.children}
                </Container>
            </MuiThemeProvider>
        );
    }
}

export { BasePageComponent }