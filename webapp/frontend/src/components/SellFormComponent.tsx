import React, { ReactNode } from 'react';

import {
  Typography,
  TextField,
  Button,
  createStyles,
  Theme,
  WithStyles,
} from '@material-ui/core';
import ItemImageUploadComponent from '../components/ItemImageUploadComponent';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import validator from 'validator';
import { ErrorMessageComponent } from './ErrorMessageComponent';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import Select from '@material-ui/core/Select';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    title: {
      marginBottom: theme.spacing(2),
    },
    form: {
      width: '80%',
      marginTop: theme.spacing(1),
    },
    selectForm: {
      minWidth: '200px',
      margin: theme.spacing(1, 0, 2),
    },
    submit: {
      margin: theme.spacing(3, 0, 2),
    },
  });

interface SellFormComponentProps extends WithStyles<typeof styles> {
  sellItem: (
    name: string,
    description: string,
    price: number,
    categoryId: number,
    image: Blob,
  ) => void;
  error?: string;
  categories: {
    id: number;
    categoryName: string;
  }[];
}

interface SellFormComponentState {
  name: string;
  description: string;
  price: number;
  selectedCategoryId: number;
  image?: Blob;
  categoryError?: string;
}

class SellFormComponent extends React.Component<
  SellFormComponentProps,
  SellFormComponentState
> {
  constructor(props: SellFormComponentProps) {
    super(props);

    this.state = {
      name: '',
      description: '',
      price: 0,
      selectedCategoryId: 0,
    };

    this._onSubmit = this._onSubmit.bind(this);
    this._onImageChange = this._onImageChange.bind(this);
    this._onChangeName = this._onChangeName.bind(this);
    this._onChangeDescription = this._onChangeDescription.bind(this);
    this._onChangeCategory = this._onChangeCategory.bind(this);
    this._onChangePrice = this._onChangePrice.bind(this);
  }

  _onSubmit(e: React.MouseEvent) {
    e.preventDefault();
    const { name, description, price, selectedCategoryId, image } = this.state;

    if (!selectedCategoryId) {
      this.setState({
        categoryError: 'カテゴリを選択してください',
      });
      return;
    }

    if (image === undefined) {
      this.setState({
        categoryError: '画像をアップロードしてください',
      });
      return;
    }

    this.props.sellItem(name, description, price, selectedCategoryId, image);
  }

  _onImageChange(image: Blob) {
    this.setState({
      image,
    });
  }

  _onChangeName(e: React.ChangeEvent<HTMLInputElement>) {
    this.setState({
      name: e.target.value,
    });
  }

  _onChangeDescription(e: React.ChangeEvent<HTMLInputElement>) {
    this.setState({
      description: e.target.value,
    });
  }

  _onChangeCategory(e: React.ChangeEvent<any>, child: ReactNode) {
    this.setState({
      selectedCategoryId: Number(e.target.value),
    });
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
    const { classes, categories } = this.props;
    const {
      name,
      description,
      price,
      selectedCategoryId,
      categoryError,
    } = this.state;

    return (
      <React.Fragment>
        <Typography className={classes.title} component="h1" variant="h5">
          出品ページ
        </Typography>
        <form className={classes.form} noValidate>
          <ItemImageUploadComponent onImageChange={this._onImageChange} />

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

          <InputLabel htmlFor="choose-category">カテゴリ</InputLabel>
          <Select
            className={classes.selectForm}
            value={selectedCategoryId}
            onChange={this._onChangeCategory}
            inputProps={{
              name: 'category',
              id: 'choose-category',
            }}
          >
            <MenuItem value={0}>
              <em>-</em>
            </MenuItem>
            {categories.map(category => (
              <MenuItem value={category.id}>{category.categoryName}</MenuItem>
            ))}
          </Select>
          {categoryError && (
            <ErrorMessageComponent id="choose-category" error={categoryError} />
          )}
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
          {this.props.error && (
            <ErrorMessageComponent id="sellButton" error={this.props.error} />
          )}
          <Button
            id="sellButton"
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

export default withStyles(styles)(SellFormComponent);
