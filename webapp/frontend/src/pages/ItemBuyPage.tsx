import React from 'react';
import {withBaseComponent} from "../hoc/withBaseComponent";
import {ItemData} from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import Typography from "@material-ui/core/Typography/Typography";
import TextField from "@material-ui/core/TextField/TextField";
import Button from "@material-ui/core/Button/Button";
import {BuyFormErrorState} from "../reducers/formErrorReducer";
import {ErrorMessageComponent} from "../components/ErrorMessageComponent";

const useStyles = makeStyles(theme => ({
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
}));

interface ItemBuyPageProps {
    item: ItemData,
    errors: BuyFormErrorState,
}

const ItemBuyPage: React.FC/*<ItemBuyPageProps>*/ = (/*{ item, errors }*/) => {
    const errors = {
        cardError: [],
        buyError: [],
    };
    const item = {
        id: 1,
        name: 'いす',
        price: 10000,
        description: 'いすです',
        createdAt: '2日前',
        thumbnailUrl: 'https://i.gyazo.com/c61ab08bca188410e81dbdcf7684e07e.png',
    };

    const classes = useStyles();
    const hasCardError = errors.cardError.length > 0;
    const hasAppError = errors.buyError.length > 0;

    return (
        <React.Fragment>
            <img className={classes.itemImage} alt={item.name} src={item.thumbnailUrl} />
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
};

export default withBaseComponent(ItemBuyPage);