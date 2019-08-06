import React from 'react';

import { makeStyles } from '@material-ui/core';
import SellFormContainer from "../containers/SellFormContainer";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";
import {BasePageComponent} from "../components/BasePageComponent";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
}));

type Props = {} & ErrorProps;

const SellPage: React.FC<Props> = () => {
    const classes = useStyles();

    return (
        <BasePageComponent>
            <div className={classes.paper}>
                <SellFormContainer />
            </div>
        </BasePageComponent>
    );
};

export default PageComponentWithError<Props>()(SellPage);
