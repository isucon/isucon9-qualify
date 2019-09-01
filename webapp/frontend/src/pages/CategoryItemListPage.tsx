import React from 'react';
import { TimelineItem } from '../dataObjects/item';
import { ItemList } from '../components/ItemList';
import SellingButtonContainer from '../containers/SellingButtonContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import BasePageContainer from '../containers/BasePageContainer';
import { createStyles, Theme, WithStyles } from '@material-ui/core';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import LoadingComponent from '../components/LoadingComponent';
import Typography from '@material-ui/core/Typography/Typography';
import { RouteComponentProps } from 'react-router';
import { InternalServerErrorPage } from './error/InternalServerErrorPage';
import validator from 'validator';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    root: {
      display: 'flex',
      flexWrap: 'wrap',
      marginTop: theme.spacing(1),
      justifyContent: 'space-around',
      overflow: 'hidden',
    },
    title: {
      paddingBottom: theme.spacing(2),
      fontWeight: theme.typography.fontWeightBold,
    },
  });

interface CategoryItemListPageProps extends WithStyles<typeof styles> {
  loading: boolean;
  load: (categoryId: number) => void;
  items: TimelineItem[];
  hasNext: boolean;
  categoryId: number;
  categoryName: string;
  loadMore: (
    createdAt: number,
    itemId: number,
    categoryId: number,
    page: number,
  ) => void;
}

type Props = CategoryItemListPageProps &
  RouteComponentProps<{ category_id: string }> &
  ErrorProps;

type State = {
  loading: boolean;
  categoryIdIsValid: boolean;
  currentCategoryId: number;
};

class CategoryItemListPage extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);

    const categoryId = this.props.match.params.category_id;
    const categoryIdIsValid = validator.isNumeric(categoryId);

    if (categoryIdIsValid) {
      this.props.load(Number(categoryId));
    }

    this.state = {
      loading: this.props.loading,
      categoryIdIsValid,
      currentCategoryId: Number(categoryId),
    };
  }

  static getDerivedStateFromProps(nextProps: Props, prevState: State) {
    const nextLoading = nextProps.loading;
    const nextCategoryId = Number(nextProps.match.params.category_id);

    // ページ遷移を確認した場合はデータ取得を行う
    if (nextCategoryId !== prevState.currentCategoryId) {
      nextProps.load(nextCategoryId);

      return {
        ...prevState,
        loading: true,
        currentCategoryId: nextCategoryId,
      };
    }

    return {
      ...prevState,
      loading: nextLoading,
      currentCategoryId: nextCategoryId,
    };
  }

  render() {
    const { classes, items, categoryName, loadMore, hasNext } = this.props;
    const { loading, currentCategoryId: categoryId } = this.state;
    const { categoryIdIsValid } = this.state;

    const Content: React.FC<{}> = () => {
      if (items.length === 0) {
        return (
          <div className={classes.root}>
            <Typography variant="h5">出品されている商品はありません</Typography>
            <SellingButtonContainer />
          </div>
        );
      }

      const lastItem = items[items.length - 1];
      const loadMoreItems = loadMore.bind(
        null,
        lastItem.createdAt,
        lastItem.id,
        categoryId,
      );

      return (
        <div className={classes.root}>
          <Typography variant="h6" className={classes.title}>
            「{categoryName}」カテゴリの新着商品一覧
          </Typography>
          <ItemList items={items} hasNext={hasNext} loadMore={loadMoreItems} />
          <SellingButtonContainer />
        </div>
      );
    };

    return (
      <BasePageContainer>
        {!categoryIdIsValid ? (
          <InternalServerErrorPage message="Category IDは数字のみです" />
        ) : loading ? (
          <LoadingComponent />
        ) : (
          <Content />
        )}
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<any>()(
  withStyles(styles)(CategoryItemListPage),
);
