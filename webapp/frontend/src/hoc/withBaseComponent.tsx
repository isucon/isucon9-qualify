import {
  ErrorType,
  InternalServerError,
  NotFoundError,
} from '../reducers/errorReducer';
import { branch, renderComponent, withProps, compose } from 'recompose';
import NotFoundContainer from '../containers/NotFoundContainer';
import InternalServerContainer from '../containers/InternalServerContainer';

export interface ErrorProps {
  errorType: ErrorType;
}

type BaseProps = ErrorProps;

export const PageComponentWithError = <Props extends ErrorProps>() =>
  compose<Props, Props>(
    withProps((props: Props) => ({
      errorType: props.errorType,
    })),
    branch<BaseProps>(
      (props: BaseProps) => props.errorType === NotFoundError,
      renderComponent(NotFoundContainer),
    ),
    branch<BaseProps>(
      (props: BaseProps) => props.errorType === InternalServerError,
      renderComponent(InternalServerContainer),
    ),
  );
