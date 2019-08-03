import React from 'react';
import {ItemData} from "../dataObjects/item";
import Typography from "@material-ui/core/Typography/Typography";
import TextField from "@material-ui/core/TextField/TextField";
import Button from "@material-ui/core/Button/Button";
import {BuyFormErrorState} from "../reducers/formErrorReducer";
import {ErrorMessageComponent} from './ErrorMessageComponent';
import {createStyles, StyleRules, Theme, WithStyles} from "@material-ui/core/styles";
import withStyles from "@material-ui/core/styles/withStyles";
import validator from 'validator';

const styles = (theme: Theme): StyleRules => createStyles({
    itemImage: {
        width: '100%',
        maxWidth: '500px',
        height: 'auto',
    },
    form: {
        width: '100%',
        marginTop: theme.spacing(1),
    },
    submit: {
        margin: theme.spacing(3, 0, 2),
    },
});

interface ItemBuyFormProps extends WithStyles<typeof styles> {
    item: ItemData,
    errors: BuyFormErrorState,
}

interface ItemBuyFormState {
    cardNumber: string,
}

class ItemBuyFormComponent extends React.Component<ItemBuyFormProps, ItemBuyFormState> {
    constructor(props: ItemBuyFormProps) {
        super(props);

        this.state = {
            cardNumber: '',
        };

        this._onChangeCardNumber = this._onChangeCardNumber.bind(this);
    }

    _onChangeCardNumber(e: React.ChangeEvent<HTMLInputElement>) {
        const cardNumber = e.target.value;

        if (cardNumber.length > 8) {
            return;
        }

        if (!validator.isHexadecimal(cardNumber) && cardNumber !== '') {
            return;
        }

        this.setState({
            cardNumber: cardNumber.toUpperCase(),
        });
    }

    render() {
        const { item, errors, classes } = this.props;
        const hasCardError = errors.cardError.length > 0;
        const hasAppError = errors.buyError.length > 0;

        return (
            <React.Fragment>
                <img className={classes.itemImage} alt={item.name} src={item.thumbnailUrl}/>
                <Typography variant="h4">{item.name}</Typography>
                <Typography variant="h5">{`¥${item.price}`}</Typography>
                <form className={classes.form} noValidate>
                    <TextField
                        variant="outlined"
                        margin="normal"
                        required
                        fullWidth
                        id="cardNumber"
                        label="カード番号"
                        name="cardNumber"
                        helperText="16進数大文字で入力してください"
                        value={this.state.cardNumber}
                        onChange={this._onChangeCardNumber}
                        autoFocus
                        FormHelperTextProps={{
                            id: 'cardNumber',
                            error: hasCardError,
                        }}
                    />
                    {
                        hasAppError &&
                        <ErrorMessageComponent errMsg={errors.buyError}/>
                    }
                    <Button
                        type="submit"
                        fullWidth
                        variant="contained"
                        color="primary"
                        className={classes.submit}
                    >
                        購入
                    </Button>
                </form>
            </React.Fragment>
        );
    }
}

export default withStyles(styles)(ItemBuyFormComponent);
