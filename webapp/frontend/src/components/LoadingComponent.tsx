import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import CircularProgress from '@material-ui/core/CircularProgress';

const useStyles = makeStyles(theme => ({
    progress: {
        margin: theme.spacing(2),
    },
}));

export default function LoadingComponent() {
    const classes = useStyles();

    return (
        <React.Fragment>
            <CircularProgress className={classes.progress} />
        </React.Fragment>
);
