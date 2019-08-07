import React, {PropsWithChildren} from 'react';

import { Container, MuiThemeProvider, createMuiTheme } from '@material-ui/core';
import LoadingComponent from "./LoadingComponent";

const themeInstance = createMuiTheme({
    palette: {
        background: {
            default: 'white'
        },
    },
});

export type Props = PropsWithChildren<{
    load: () => void;
    alreadyLoaded: boolean,
    loading: boolean,
}>

class BasePageComponent extends React.Component<Props> {
    constructor(props: Props) {
        super(props);

        if (!this.props.alreadyLoaded) {
            this.props.load();
        }
    }

    render() {
        return (
            <MuiThemeProvider theme={themeInstance}>
                <Container maxWidth="lg">
                    {
                        this.props.loading ? (
                            <LoadingComponent />
                        ) : (
                            this.props.children || null
                        )
                    }
                </Container>
            </MuiThemeProvider>
        );
    }
}

export { BasePageComponent }