import React from 'react';
import { ItemData } from '../dataObjects/item';
import Typography from '@material-ui/core/Typography/Typography';
import TextField from '@material-ui/core/TextField/TextField';
import { BuyFormErrorState } from '../reducers/formErrorReducer';
import { ErrorMessageComponent } from './ErrorMessageComponent';
import {
  createStyles,
  StyleRules,
  Theme,
  WithStyles,
} from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import validator from 'validator';
import { LoadingButton } from './LoadingButton';
import { Link } from 'react-router-dom';
import { routes } from '../routes/Route';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    itemImage: {
      width: '100%',
      maxWidth: '500px',
      height: 'auto',
      textAlign: 'center',
    },
    form: {
      width: '100%',
      marginTop: theme.spacing(3, 0, 1),
    },
  });

interface ItemBuyFormProps extends WithStyles<typeof styles> {
  item?: ItemData;
  onBuyAction: (itemId: number, cardNumber: string) => void;
  loadingBuy: boolean;
  errors?: BuyFormErrorState;
}

interface ItemBuyFormState {
  cardNumber: string;
}

class ItemBuyFormComponent extends React.Component<
  ItemBuyFormProps,
  ItemBuyFormState
> {
  constructor(props: ItemBuyFormProps) {
    super(props);

    this.state = {
      cardNumber: '',
    };

    this._onChangeCardNumber = this._onChangeCardNumber.bind(this);
    this._onClickBuyButton = this._onClickBuyButton.bind(this);
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

  _onClickBuyButton(e: React.MouseEvent) {
    const { item } = this.props;
    const id = item ? item.id : 0; // MEMO: この関数がcallされる前にitemの有無は確認されるので0になることはない
    const { cardNumber } = this.state;
    this.props.onBuyAction(id, cardNumber);
  }

  render() {
    const { item, errors, classes, loadingBuy } = this.props;
    const cardError = errors ? errors.cardError : undefined;
    const appError = errors ? errors.buyError : undefined;

    if (!item) {
      return (
        <React.Fragment>
          <Typography variant="h4">エラーが発生しました</Typography>
          <Link to={routes.timeline.path}>トップページへ戻る</Link>
        </React.Fragment>
      );
    }

    return (
      <React.Fragment>
        <img
          className={classes.itemImage}
          alt={item.name}
          src={item.thumbnailUrl}
        />
        <Typography variant="h4">{item.name}</Typography>
        <Typography variant="h5">{`${item.price}ｲｽｺｲﾝ`}</Typography>
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
            error={!!cardError}
            value={this.state.cardNumber}
            onChange={this._onChangeCardNumber}
            autoFocus
          />
          {cardError && (
            <ErrorMessageComponent id="cardNumber" error={cardError} />
          )}
          <LoadingButton
            onClick={this._onClickBuyButton}
            buttonName={'購入'}
            loading={loadingBuy}
          />
          {appError && (
            <ErrorMessageComponent id="buyButton" error={appError} />
          )}
        </form>
      </React.Fragment>
    );
  }
}

export default withStyles(styles)(ItemBuyFormComponent);
