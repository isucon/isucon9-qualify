import React from 'react';
import BasePageContainer from '../containers/BasePageContainer';
import {
  Button,
  createStyles,
  TextField,
  Theme,
  Typography,
  WithStyles,
} from '@material-ui/core';
import { StyleRules } from '@material-ui/core/styles';
import { ItemData } from '../dataObjects/item';
import { RouteComponentProps } from 'react-router';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import withStyles from '@material-ui/core/styles/withStyles';
import LoadingComponent from '../components/LoadingComponent';
import { ErrorMessageComponent } from '../components/ErrorMessageComponent';
import validator from 'validator';
import { InternalServerErrorPage } from './error/InternalServerErrorPage';
import { Link as RouteLink } from 'react-router-dom';
import { routes } from '../routes/Route';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    link: {
      textDecoration: 'none',
    },
  });

interface BaseProps extends WithStyles<typeof styles> {
  loading: boolean;
  load: (itemId: number) => void;
  item?: ItemData;
  formError?: string;
  onClickEdit: (itemId: number, price: number) => void;
}

type Props = BaseProps & RouteComponentProps<{ item_id: string }> & ErrorProps;

interface State {
  price: number;
}

class ItemEditPage extends React.Component<Props, State> {
  private readonly itemId: number;

  constructor(props: Props) {
    super(props);

    this.itemId = Number(this.props.match.params.item_id);
    this.props.load(this.itemId);

    const { item } = this.props;

    this.state = {
      price: item ? item.price : 0,
    };

    this._onClickEdit = this._onClickEdit.bind(this);
    this._onChangePrice = this._onChangePrice.bind(this);
  }

  _onClickEdit(e: React.MouseEvent) {
    e.preventDefault();
    this.props.onClickEdit(this.itemId, this.state.price);
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
    });
  }

  render() {
    const { loading, item, formError, classes } = this.props;
    const { price } = this.state;

    if (loading) {
      return <LoadingComponent />;
    }

    if (!item) {
      return <InternalServerErrorPage message="商品が読み込めませんでした" />;
    }

    return (
      <BasePageContainer>
        <Typography component="h1" variant="h5">
          商品編集ページ
        </Typography>
        <Typography component="h2">{item.name}</Typography>
        <form className={classes.form} noValidate>
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
          {formError && (
            <ErrorMessageComponent id="sellButton" error={formError} />
          )}
          <Button
            id="editButton"
            type="submit"
            fullWidth
            variant="contained"
            color="primary"
            className={classes.submit}
            onClick={this._onClickEdit}
          >
            編集する
          </Button>
          <RouteLink to={routes.item.getPath(item.id)}>
            商品ページへ戻る
          </RouteLink>
        </form>
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<any>()(withStyles(styles)(ItemEditPage));
