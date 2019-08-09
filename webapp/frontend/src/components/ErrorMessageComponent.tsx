import React from 'react';
import { FormHelperText } from '@material-ui/core';

interface ErrorMessageComponentProps {
  id: string;
  error: string;
}

const ErrorMessageComponent: React.FC<ErrorMessageComponentProps> = ({
  id,
  error,
}) => {
  return (
    <FormHelperText key={error} id={id} error={true}>
      {error}
    </FormHelperText>
  );
};

export { ErrorMessageComponent };
