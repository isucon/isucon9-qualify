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
  categoryIdIsValid: boolean;
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
      categoryIdIsValid,
    };
  }

  render() {
    const {
      classes,
      loading,
      items,
      categoryId,
      categoryName,
      loadMore,
      hasNext,
    } = this.props;
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
          <Typography variant="h6">{categoryName}の新着商品</Typography>
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
