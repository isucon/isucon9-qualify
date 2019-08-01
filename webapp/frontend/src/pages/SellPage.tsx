import React from 'react';

import { Typography, TextField, Button, makeStyles } from '@material-ui/core';
import ItemImageUploadComponent from "../components/ItemImageUploadComponent";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    form: {
        width: '80%',
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
}));

const SellPage: React.FC = () => {
    const classes = useStyles();

    return (
        <div className={classes.paper}>
            <Typography component="h1" variant="h5">
                出品ページ
            </Typography>
            <form className={classes.form} noValidate>
                <ItemImageUploadComponent />

                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="name"
                    label="商品名"
                    name="name"
                    autoFocus
                />
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="description"
                    label="商品説明"
                    name="name"
                    multiline
                />
                <TextField
                    variant="outlined"
                    margin="normal"
                    required
                    fullWidth
                    id="price"
                    label="値段"
                    name="price"
                />
                <Button
                    type="submit"
                    fullWidth
                    variant="contained"
                    color="primary"
                    className={classes.submit}
                >
                    出品する
                </Button>
            </form>
        </div>
    );
};

export { SellPage }
