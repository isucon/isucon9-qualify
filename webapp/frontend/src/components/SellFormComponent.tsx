import React from 'react';

import {Typography, TextField, Button, createStyles, Theme, WithStyles} from '@material-ui/core';
import ItemImageUploadComponent from "../components/ItemImageUploadComponent";
import {StyleRules} from "@material-ui/core/styles";
import withStyles from "@material-ui/core/styles/withStyles";
import validator from 'validator';

const styles = (theme: Theme): StyleRules => createStyles({
    form: {
        width: '80%',
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
});

interface SellFormComponentProps extends WithStyles<typeof styles> {
    sellItem: (name: string, description: string, price: number) => void
}

interface SellFormComponentState {
    name: string,
    description: string,
    price: number,
}

class SellFormComponent extends React.Component<SellFormComponentProps, SellFormComponentState> {
    constructor(props: SellFormComponentProps) {
        super(props);

        this.state = {
            name: '',
            description: '',
            price: 0,
        };

        this._onSubmit = this._onSubmit.bind(this);
        this._onChangeName = this._onChangeName.bind(this);
        this._onChangeDescription = this._onChangeDescription.bind(this);
        this._onChangePrice = this._onChangePrice.bind(this);
    }

    _onSubmit(e: React.MouseEvent) {
        e.preventDefault();
        const { name, description, price } = this.state;
        this.props.sellItem(name, description, price);
    }

    _onChangeName(e: React.ChangeEvent<HTMLInputElement>) {
        this.setState({
            name: e.target.value
        })
    }

    _onChangeDescription(e: React.ChangeEvent<HTMLInputElement>) {
        this.setState({
            description: e.target.value
        })
    }

    _onChangePrice(e: React.ChangeEvent<HTMLInputElement>) {
        const input = e.target.value;

        // Only allow number
        if (!validator.isNumeric(input, { no_symbols: true })) {
            e.preventDefault();
            return;
        }

        this.setState({
            price: Number(input),
        })
    }

    render() {
        const { classes } = this.props;
        const { name, description, price } = this.state;

        return (
            <React.Fragment>
                <Typography component="h1" variant="h5">
                    出品ページ
                </Typography>
                <form className={classes.form} noValidate>
                    <ItemImageUploadComponent/>

                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        id="name"
                        label="商品名"
                        name="name"
                        value={name}
                        onChange={this._onChangeName}
                        autoFocus
                    />
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        id="description"
                        label="商品説明"
                        name="description"
                        value={description}
                        onChange={this._onChangeDescription}
                        multiline
                        rows={5}
                    />
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        id="price"
                        label="値段"
                        name="price"
                        value={price}
                        onChange={this._onChangePrice}
                    />
                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        color="primary"
                        className={classes.submit}
                        onClick={this._onSubmit}
                    >
                        出品する
                    </Button>
                </form>
            </React.Fragment>
        );
    }
}

export default withStyles(styles)(SellFormComponent)
