import React from 'react';
import Fab from "@material-ui/core/Fab/Fab";
import Icon from "@material-ui/core/Icon/Icon";
import makeStyles from "@material-ui/core/styles/makeStyles";

const useStyles = makeStyles(theme => ({
    fab: {
        margin: theme.spacing(1),
    },
}));

interface ListingButtonComponentProps {
    onClick: Function
}

const ListingButtonComponent: React.FC<ListingButtonComponentProps> = ({ onClick }) => {
    const classes = useStyles();

    return (
        <Fab className={classes.fab} size="medium" color="secondary">
            <Icon>edit_icon</Icon>
        </Fab>
    );
};

export { ListingButtonComponent }